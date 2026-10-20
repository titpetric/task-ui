package assert_test

import (
	"errors"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// recordingTB stands in for a *testing.T so a test can call the failure
// reporters without failing itself. The reporters reach only Fail and
// FailNow; nothing touches the embedded nil TB.
type recordingTB struct {
	testing.TB

	failed bool
}

func (r *recordingTB) Fail() {
	r.failed = true
}

func (r *recordingTB) FailNow() {
	r.failed = true
}

// mustFail calls a failing assertion against a recordingTB, so the recordingTB
// records the failure and the test that drives it passes. Every assertion prints
// its message to stdout when it fails, so a passing run prints one line per
// case. Those lines show the reporter working.
func mustFail(t *testing.T, name string, failing func(testing.TB)) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		tb := &recordingTB{}
		failing(tb)
		assert.True(t, tb.failed, name+" must fail the TB it is given")
	})
}

// suiteFixture exercises TestSuite and Run.
type suiteFixture struct {
	assert.TestSuite

	setup    int
	tearDown int
	ran      int
}

// bareSuite has neither SetupTest nor TearDownTest, which is the branch of
// Run that suiteFixture does not reach.
type bareSuite struct {
	assert.TestSuite

	ran int
}

func (s *bareSuite) TestBare() {
	s.ran++
}

func (s *suiteFixture) SetupTest() {
	s.setup++
}

func (s *suiteFixture) TearDownTest() {
	s.tearDown++
}

func (s *suiteFixture) TestPasses() {
	s.ran++
	assert.True(s.T(), true, "the suite hands its T to the test")
}

// TestAssert invokes every assertion on the path where it holds.
func TestAssert(t *testing.T) {
	t.Run("core", assertCore)
	t.Run("reporters", assertReporters)
	t.Run("suite", assertSuite)
	t.Run("extra", assertExtra)
	t.Run("atkins", assertAtkins)
}

// TestAssertFailure invokes every assertion on the path where it does not
// hold, which is the branch that raises the failure.
func TestAssertFailure(t *testing.T) {
	t.Run("core", assertCoreFailure)
	t.Run("extra", assertExtraFailure)
	t.Run("atkins", assertAtkinsFailure)
}

func assertCore(t *testing.T) {
	assert.True(t, assert.ObjectsAreEqualValues(1, 1))
	assert.False(t, assert.ObjectsAreEqualValues(1, 2))

	assert.Equal(t, "a", "a", "strings match")
	assert.EqualValues(t, []int{1, 2}, []int{1, 2})
	assert.NotEqual(t, 1, 2)

	assert.NoError(t, nil)
	assert.Error(t, errors.New("boom"))

	assert.True(t, true)
	assert.False(t, false)

	assert.Contains(t, "haystack", "stack")
	assert.Contains(t, []string{"a", "b"}, "b")
	assert.Contains(t, map[string]int{"a": 1}, "a")

	assert.CheckEquals(t, 42, 42, "integers match")
	assert.Assert(t, true, "condition holds for %d", 42)
}

// assertReporters drives Errorf and Fail against a stand-in TB, so the stand-in
// records the failure they raise and this test passes. Both print the
// message to stdout: those are the two lines in a passing run.
func assertReporters(t *testing.T) {
	errorf := &recordingTB{}
	assert.Errorf(errorf, "reporter check: %d", 42)
	assert.True(t, errorf.failed, "Errorf fails the TB it is given")

	fail := &recordingTB{}
	assert.Fail(fail, "reporter check", "fail")
	assert.True(t, fail.failed, "Fail fails the TB it is given")
}

func assertSuite(t *testing.T) {
	fixture := &suiteFixture{}
	fixture.SetT(t)
	assert.True(t, fixture.T() == t, "SetT and T agree")

	assert.Run(t, fixture)
	assert.Equal(t, 1, fixture.setup)
	assert.Equal(t, 1, fixture.ran)
	assert.Equal(t, 1, fixture.tearDown)

	// A suite with neither hook takes the other branch of Run.
	bare := &bareSuite{}
	assert.Run(t, bare)
	assert.Equal(t, 1, bare.ran)
}

// assertCoreFailure drives the failing path of every assertion in assert.go.
func assertCoreFailure(t *testing.T) {
	mustFail(t, "NotEqual", func(tb testing.TB) { assert.NotEqual(tb, 1, 1) })
	mustFail(t, "EqualValues", func(tb testing.TB) { assert.EqualValues(tb, []int{1}, []int{2}) })
	mustFail(t, "Equal", func(tb testing.TB) { assert.Equal(tb, "a", "b") })

	mustFail(t, "NoError", func(tb testing.TB) { assert.NoError(tb, errors.New("boom")) })
	mustFail(t, "Error", func(tb testing.TB) { assert.Error(tb, nil) })

	mustFail(t, "True", func(tb testing.TB) { assert.True(tb, false) })
	mustFail(t, "False", func(tb testing.TB) { assert.False(tb, true) })

	mustFail(t, "Contains", func(tb testing.TB) { assert.Contains(tb, "haystack", "needle") })
	// A haystack reflect cannot read at all takes the invalid branch.
	mustFail(t, "ContainsInvalid", func(tb testing.TB) { assert.Contains(tb, nil, "needle") })

	mustFail(t, "CheckEquals", func(tb testing.TB) { assert.CheckEquals(tb, 41, 42, "integers differ") })
	mustFail(t, "Assert", func(tb testing.TB) { assert.Assert(tb, false, "condition fails for %d", 42) })
}
