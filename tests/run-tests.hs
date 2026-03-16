#!/usr/bin/env runhaskell

{-# LANGUAGE OverloadedStrings #-}

import Control.Applicative (asum)
import Control.Arrow ((&&&))
import Control.Monad (unless)
import Data.List (dropWhileEnd, find, isInfixOf, isPrefixOf, isSuffixOf, partition, sort)
import Data.Maybe (catMaybes, fromMaybe, isJust)
import Data.Text (Text)
import Data.Text qualified as Text
import Data.Text.IO qualified as Text
import System.Directory (listDirectory)
import System.Environment (getArgs)
import System.Exit (ExitCode (..), exitWith)
import System.FilePath (takeFileName, (</>))
import System.Process (readProcessWithExitCode)

data TestType = Passing | Failing

main :: IO ()
main = do
  args <- getArgs
  let exe = case args of
        [e] -> e
        _ -> error "Usage: run-tests.sh <executable>"

  let testDir = "tests"

  files <- listDirectory testDir
  let (passFiles, failFiles) =
        partition ("pass-" `isInfixOf`)
          . map (testDir </>)
          . filter (".co" `isSuffixOf`)
          $ files

  (passed, failed, errors) <- runTests exe passFiles failFiles

  putStrLn ""
  putStrLn $ "Results: " <> show passed <> " passed, " <> show failed <> " failed"

  unless (null errors) $ do
    putStrLn ""
    putStrLn "Failures:"
    mapM_ (Text.putStrLn . ("  - " <>)) errors
    exitWith (ExitFailure 1)

runTests :: FilePath -> [FilePath] -> [FilePath] -> IO (Int, Int, [Text])
runTests exe passFiles failFiles = do
  passResults <- mapM (testPassing exe) passFiles
  failResults <- mapM (testFailing exe) failFiles

  let passed = length (filter snd passResults) + length (filter snd failResults)
      failed = length (filter (not . snd) passResults) + length (filter (not . snd) failResults)
      errors = concatMap fst (filter (not . snd) passResults) <> concatMap fst (filter (not . snd) failResults)

  return (passed, failed, errors)

testPassing :: FilePath -> FilePath -> IO ([Text], Bool)
testPassing exe f = do
  let name = takeFileName f
  content <- Text.readFile f
  let hasNoOutput = isJust $ find ("// nooutput" `Text.isPrefixOf`) $ Text.lines content
      hasOutput = not hasNoOutput
  let expectedOutput = extractExpectedOutput content
  (exitCode, stdout, _) <- readProcessWithExitCode exe [f] ""
  let actualOutput = Text.pack stdout
  case exitCode of
    ExitSuccess | hasOutput && Text.null actualOutput -> do
      putStrLn $ "  FAIL  " <> name <> " (expected output but got none)"
      return ([Text.pack name <> ": expected output but got none"], False)
    ExitSuccess | hasNoOutput && Text.null actualOutput -> do
      putStrLn $ "  PASS  " <> name
      return ([], True)
    ExitSuccess | hasNoOutput -> do
      putStrLn $ "  FAIL  " <> name <> " (expected no output but got output)"
      return ([Text.pack name <> ": expected no output but got: " <> actualOutput], False)
    ExitSuccess | actualOutput == expectedOutput -> do
      putStrLn $ "  PASS  " <> name
      return ([], True)
    ExitSuccess -> do
      putStrLn $ "  FAIL  " <> name <> " (wrong output)"
      return ([Text.pack name <> ": expected output:\n" <> expectedOutput <> "\nbut got:\n" <> actualOutput], False)
    ExitFailure _ -> do
      putStrLn $ "  FAIL  " <> name <> " (expected to pass)"
      return ([Text.pack name <> ": expected to pass but failed"], False)

extractExpectedOutput :: Text -> Text
extractExpectedOutput =
  Text.unlines
    . map (flip (foldl' (\h (n, r) -> Text.replace n r h)) escapes)
    . concatMap (Text.splitOn ";")
    . map (Text.strip . Text.takeWhileEnd (/= '/'))
    . filter (\l -> "print" `Text.isInfixOf` l && "//" `Text.isInfixOf` l)
    . Text.lines
  where
    escapes =
      [ ("\\n", "\n"),
        ("\\t", "\t"),
        ("\\r", "\r"),
        ("\\\"", "\""),
        ("\\\\", "\\")
      ]

testFailing :: String -> FilePath -> IO ([Text], Bool)
testFailing exe f = do
  let name = takeFileName f
  content <- Text.readFile f
  let expected = extractExpectedError content
  (exitCode, _, actualS) <- readProcessWithExitCode exe [f] ""
  let actual = Text.pack actualS
  case exitCode of
    ExitSuccess -> do
      putStrLn $ "  FAIL  " <> name <> " (expected to fail)"
      return ([Text.pack name <> ": expected to fail but passed"], False)
    ExitFailure _ | expected `Text.isInfixOf` actual -> do
      putStrLn $ "  PASS  " <> name
      return ([], True)
    ExitFailure _ -> do
      putStrLn $ "  FAIL  " <> name <> " (wrong error)"
      return ([Text.pack name <> ": expected '" <> expected <> "' but got '" <> actual <> "'"], False)

extractExpectedError :: Text -> Text
extractExpectedError =
  fromMaybe ""
    . asum
    . map (Text.stripPrefix "// ERROR: ")
    . Text.lines
