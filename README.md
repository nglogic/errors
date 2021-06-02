# Errors

`errors` is a go package that defines the error type, which carries type information and some extra values for usage across API boundaries.

`errors` is designed to work nicely with standard go errors. It can be a drop-in replacement for the standard errors package, or it can be mixed with it in any way. You can use `fmt.Errorf("%w")`, `errors.Is` and `errors.As` safely with it, just like standard errors.

The package also provides `Append` function for representing a list of errors as a single error.

## But why?

There is a standard way of passing context data down the call stack in go: the `context.Context`. It's used to carry deadlines and request scoped data. It's often used to carry some extra information across the API boundaries. For example: you can have a handler that stores request id in the context. Then, down the stack, if you have to log something, you can log with the id extracted from the context (see the following diagram, left side goes through the stack with a context). So there's a way to communicate between the handler layer and app layer.

But what if you want to pass information the other way in case of a failure? Look at the following diagram (right side goes back the stack and caries an error). The only thing you have is an error.

![diagram](docs/errors.svg)

The usual solution is to have a few sentinel errors, and choose the action by the result of comparing the errors to predefined sentinels by `errors.Is`. It works, but isn't very flexible. For example:

- You want to have a dynamic error message. Sentinel errors don't allow it.
- You want to return a specific message to the user, but also log all the error details internally.

This package gives you the tools to do these things.

## Types

The type describes a general error category. It is similar to HTTP status classes (like `4xx` or `5xx`).

The idea is that an error can carry a type across API boundaries. You can use this information to decide what is the proper way of handling the error.

```go
err := errors.New("file not found").WithType(errors.TypeNotFound)

---

if errors.IsType(errors.TypeNotFound) { // returns true!
    // This is actually expected...
} else {
    // This is the bad case.
    log.Print(err)
}
```

### Creating custom type

There is a predefined list of the types that should be enough for the majority of use cases. 
However, if you need a different one, adding it is as simple as defining a const:

```go
const MyTypeResourceExhausted errors.Type = "ResourceExhausted"
```

## Context values

TODO

## Multi-error

TODO

## Wrapping

This package works well with standard wrapping method:

```go
err := errors.New("testing error")
err = fmt.Errorf("something bad: %w", err)
```

You can still use standard errors `Is` and `As` methods to inspect error values.
