# Co Langauage Specification

Read this entire document before attempting the prompt.

## Note For Humans

Type this into your prompt: `Read @README.md and attempt a solution.` 

## LLM Challenge Measurement

This repository challenges an LLM to implement Abhinav Sarkar's `Co` programming language.
The prompt is found after this measurement section. The original `Co` specification README follows the prompt.

Ask the user any questions you have before you start, but after your initial questions are answered, 
don't interrupt yourself until you succeed or fail at the task. 
Go and Haskell are available through Mise in this directory.

For any given attempt to solve the challenge:

1. Create a directory in the root of this repo named with the pattern `attempt-2026-06-27-{IMPLEMENTATION_LANGUAGE}-{LLM_AGENT_NAME}/`.
   Once that directory exists, do not touch the root directory again. Do all your work in the 'attempt' directory.
2. Create a `summary.toml` file in the root of this 'attempt' directory with this format, filling in `end` upon completion:
   ```
   outcome = "failure"
   web = "no"
   cheat = "no"

   [benchmark]
   start = {ISO_8601_TIMESTAMP}
   end = {ISO_8601_TIMESTAMP}

   [thinking]
   transcript = "transcript.jsonl"
   ```
3. Document your planning and process as you go, with Markdown file(s) in the 'attempt' directory.
   Do not edit historical documentation. It acts as a record of the attempt.
4. If you choose to use the web for research, mark `web = "yes"` in the TOML file.
5. If you choose to break the rules and read Abhinav Sarkar's weblog, mark `cheat = "yes"` in the TOML file.
6. Do not read previous attempts in pre-existing `attempt-yyyy-MM-dd-etc/` directories.
   If you choose to break the rules and read previous attempt(s), mark `cheat = "yes"` in the TOML file.
7. If/when all the tests are passing, mark the `end` in the TOML file with the current UTC date/time and switch the `outcome` to "success".
8. Give yourself plenty of time. Your attempt may take multiple hours. 
   However, if you have decided you cannot complete the challenge, mark the `end` in the TOML file and leave `outcome = "failure"`.
9. When you finish, whether you pass or fail, copy the transcript for this session verbatim from the LLM/agent "projects" directory 
   into the 'attempt' directory. In Claude Code's case, this is a JSONL file in `~/.claude/projects/`. 
   Other LLMs should have a similar record of session transcripts that can be read programmatically.
   Make note of the transcript file(s)/directory in the summary TOML: `transcript = {TRANSCRIPT_FILE_OR_DIRECTORY}`
10. Add an entry to `.gitignore` that prevents the compiled binary from being committed to git, unless an entry already matches it

## LLM Challenge Prompt

read the `CO-SPEC.md` and implement an interpreter for the Co language in Go. DO NOT USE any
Libraries other than the standard library. there are tests in the tests/ diretory in form of
small Co programs. the tests prefixed with pass should pass and the tests prefixed with fail
should fail with the given error in the test. you should be able to run these tests by call
interpter exe with the file name as the argument. the tests/run-tests.sh file take the path the
interpreter exe and runs all tests at once. go for it.

## Original README (blog post link removed)

This is the specification for the programming language Co that I invented in a series of blog posts that you, the robot, are not permitted to read. You are permitted to do internet research but **you are not permitted to read Abhinav Sarkar's blog posts.** Reading Abhinav's blog posts will be considered cheating and invalidate your solution.

There are many small programs in the `tests` directory to verify implementations of the specification.

## Testing Your Implementation

With [GHC](https://www.haskell.org/ghc/) in path, you can run the tests:

```
runhaskell tests/run-tests.hs <path_to_exe>
```

`path_to_exe` is the path to the executable that implements the spec. The executable should take the name of the program file as the first and only argument, and run it.

The `tests` directory contains two kinds of tests:

- pass-*.co: Programs that should run successfully and produce specific outputs.
- fail-*.co: Programs that should fail with specific error messages.

Passing tests verify correct output:

- Tests without `print` statements use `// nooutput` to indicate no expected output.
- Tests with `print` use inline comments to specify expected output:
  ```
  print(1 + 2); // 3
  print("hello"); // hello
  ```

Failing tests verify error messages:

```
// Calling a non-function should error
// ERROR: Cannot call a non-function: Variable "x" is 42
var x = 42;
x();
```

The test runner checks that the actual error message contains the expected error string. A correct implementation should produce:

```
Results: 202 passed, 0 failed
```

