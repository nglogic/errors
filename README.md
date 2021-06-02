# Errors

The errors package allows adding type and label information to application errors.

## Wrapping

This package works well with standard wrapping method:

```go
err := errors.New("testing error")
wrapped := fmt.Errorf("something bad: %w", err)
```

You can still use standard errors `Is` and `As` methods to inspect error values.
## Types and status codes

You can mark errors with predefined error types, or with your own types.

Type describes general error category. It bears information about error group, which can be "client" or "server".

```go
err := errors.New("file not found").WithType(errors.TypeNotFound)

---

if errors.IsType(errors.TypeNotFound) { // returns true!
    // This is actually expected...
} else {
    // This is the bad case.
    log.Print(err)
}

---

if errors.IsGroup(errors.GroupClient) {
    // This is not a system fault, no need to panic.
} else {
    // This is the bad case.
    log.Print(err)
}
```

### Creating your own specific type:

```go
MyErrorType := errors.Type{
    Name: "ResourceExhausted", // Name of the type.
    Group: GroupServer, // Error caused by server problem.
}
```
