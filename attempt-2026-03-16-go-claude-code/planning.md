# Co Interpreter Implementation Plan

## Architecture
- Single Go file for simplicity: `main.go`
- Phases: Lexer -> Parser -> Scope Analysis -> Interpreter
- Cooperative coroutine scheduler with channels

## Key Design Decisions
1. **Lexer**: Hand-written tokenizer with all operators, keywords, literals
2. **Parser**: Recursive descent with precedence climbing for expressions
3. **Scope Analysis**: Pre-execution pass to detect variable errors (no hoisting, redefinition rules)
4. **Interpreter**: Tree-walking interpreter with environment chains
5. **Concurrency**: Go goroutines won't work for cooperative scheduling - need manual coroutine scheduler
   - Use a priority queue keyed by time
   - Each coroutine is a goroutine that communicates with scheduler via Go channels
   - Yield/sleep/channel-block all signal scheduler

## Scope Rules (tricky parts)
- if/while do NOT create new scopes
- Variables in while bodies are "while-local" and can be redefined across iterations
- Variables in if bodies ARE in enclosing scope
- Two if blocks defining same var = error
- Function params CAN shadow outer vars
- Named functions have self-reference; lambdas do NOT
- var x = x + 1 is error (self-reference in initializer)

## Error Messages
Must match exact format from spec for test runner (substring match).

## Concurrency Model
- Scheduler with priority queue (by time)
- Main coroutine + spawned coroutines
- Channels: unbuffered and buffered
- Send/receive queues (max 4 each)
- Deadlock detection: main blocked + empty scheduler queue
