1. Error Handling
Wrapping: Always wrap errors with context using fmt.Errorf("context: %w", err).

Checking: Use errors.Is() or errors.As() for error comparisons; never compare error strings directly.

Return Early: Use the "Line of Sight" principle. Handle errors immediately and return, keeping the "happy path" aligned to the left.

### Examples
 //  Bad: String comparison and nested happy path
if err != nil && err.Error() == "connection failed" {
    return err
} else {
    // nested logic...
}

//  Good: errors.Is and Line of Sight
if errors.Is(err, ErrConnFailed) {
    return fmt.Errorf("network layer failure: %w", err)
}
// happy path continues here...

2. Naming Conventions
Receivers: Use short, 1-3 letter abbreviations (e.g., func (jb *JourneyBuilder) ... rather than func (this *JourneyBuilder) ...).

Interfaces: Name interfaces based on what they do, usually ending in "er" (e.g., Processor, Mapper).

Variables: Use camelCase for internal variables and PascalCase for exported ones.

//  Bad: Generic 'this' and wordy interface
type JourneyProcessingInterface interface {}
func (this *JourneyBuilder) Start() {}

//  Good: 'er' suffix and 2-letter receiver
type Processor interface {}
func (jb *JourneyBuilder) Start() {}

3. Concurrency & Performance
Mutexes: Always defer mu.Unlock() immediately after mu.Lock() to prevent deadlocks.

Context: Always pass context.Context as the first argument to functions performing I/O or long-running tasks.

Slices: Pre-allocate slice capacity with make([]T, 0, length) if the final size is known.

//  Bad: Manual unlock and dynamic slice growth
mu.Lock()
doSomething()
mu.Unlock() 

list := []string{}
for _, v := range input { list = append(list, v) }

//  Good: Deferred unlock and pre-allocation
mu.Lock()
defer mu.Unlock()

list := make([]string, 0, len(input))
for _, v := range input { list = append(list, v) }

4. Documentation
Exported Functions: Every exported function must have a comment starting with the function name.

Complex Logic: Use "Why, not How" comments for non-obvious business logic in the Journey engine.

//  Bad: Comment explains the 'How'
// increment the counter by one
jb.StepCount++

//  Good: Comment explains the 'Why'
// We increment here to satisfy the rate-limiter check in the next node
jb.StepCount++
