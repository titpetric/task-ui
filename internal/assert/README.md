# Assert embeddings

This package exists to drop the dependency on stretchr/testify/assert.
It mutes the conversation on which to use (assert or require), and
provides a limited API that does about the same.

It's based on [fortio.org/assert](https://github.com/fortio/assert) and
contains the following modifications:

- Use `testing.TB` instead of explicit `*testing.T`
- Use `any` or `...any` for printf arguments

It's meant to be used from inside the current module and not
intended as an imported dependency.

## Copied from the platform

`assert.go` and `assert_extra.go` are a copy of
`github.com/titpetric/platform/internal/assert`, which is not importable
across modules. Two changes were made to the copy:

- `orderedCompare` became `ordered` and returns a second value saying
  whether the two operands are comparable at all. `Greater` read a bare
  0 as "not greater" and was right by accident; `LessOrEqual` and
  `GreaterOrEqual` would have read it as "equal" and passed.
- `assert_atkins.go` adds what atkins uses and the platform did not:
  `AnError`, `Positive`, `Zero`, `NotZero`, `Less`, `LessOrEqual`,
  `GreaterOrEqual`, `FileExists`, `NoFileExists`, `DirExists`,
  `NoDirExists`, `JSONEq`, `ErrorAs`, `ElementsMatch` and `Eventually`.

## No require

There is no `require`, so a failed assertion reports and continues
instead of calling `FailNow`. A test that dereferences a value it just
asserted on panics on failure rather than stopping at the assertion.
The panic still fails the test; it reads worse in the output.
