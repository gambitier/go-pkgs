package domainerr

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	cerrors "github.com/cockroachdb/errors"
)

func TestErrorsIsMatchesByCode(t *testing.T) {
	err := NotFound("customer not found", nil, map[string]any{"customerId": "123"})

	if !errors.Is(err, &Error{Code: CodeNotFound}) {
		t.Fatalf("expected errors.Is(err, CodeNotFound) to be true")
	}
	if errors.Is(err, &Error{Code: CodeConflict}) {
		t.Fatalf("expected errors.Is(err, CodeConflict) to be false")
	}
}

func TestErrorsIsMatchesWrappedByCode(t *testing.T) {
	err := Wrap(NotFound("invoice not found", nil, map[string]any{"invoiceId": "inv_1"}), "failed to fetch invoice")

	if !errors.Is(err, &Error{Code: CodeNotFound}) {
		t.Fatalf("expected wrapped not found error to match CodeNotFound")
	}
	if errors.Is(err, &Error{Code: CodeInvalidArgument}) {
		t.Fatalf("expected wrapped not found error to not match CodeInvalidArgument")
	}
}

func TestStackAndSourceAvailableForInternal(t *testing.T) {
	err := Internal("storage failure", io.EOF, nil)
	if got := OneLineSource(err); got == "" {
		t.Fatalf("expected source to be populated")
	}
	if got := strings.TrimSpace(StackTraceString(err)); got == "" {
		t.Fatalf("expected stack trace string to be populated")
	}
}

func TestErrorContextUsesOnlyWrapMessages(t *testing.T) {
	base := Internal("logging scenario failed", errors.New("logging scenario database timeout"), nil)
	level1 := Wrap(base, "failed to execute logging scenario")
	level2 := Wrap(level1, "failed to load logging scenario")
	level3 := Wrap(level2, "failed to run logging error scenario")

	got := ErrorContext(level3)
	want := "failed to run logging error scenario: failed to load logging scenario: failed to execute logging scenario"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRootCauseReturnsDeepestLeafMessage(t *testing.T) {
	base := Internal("logging scenario failed", errors.New("logging scenario database timeout"), nil)
	level1 := Wrap(base, "failed to execute logging scenario")

	if got := RootCause(level1); got != "logging scenario database timeout" {
		t.Fatalf("expected deepest leaf cause, got %q", got)
	}
}

func TestLogFieldsExposesCauseChainForOpaqueDomainError(t *testing.T) {
	err := Internal(
		"Wasabi upload failed",
		fmt.Errorf("storage set: %w", errors.New("AccessDenied: User is not allowed to access resource")),
		nil,
	)

	fields := LogFields(err)
	if fields["error"] == nil {
		t.Fatalf("expected error field, got %#v", fields)
	}
	errorValue, _ := fields["error"].(string)
	if errorValue == "Wasabi upload failed" {
		t.Fatalf("expected error field to include cause chain, got %q", errorValue)
	}
	if !strings.Contains(errorValue, "Wasabi upload failed") {
		t.Fatalf("expected domain message in chain, got %q", errorValue)
	}
	if !strings.Contains(errorValue, "AccessDenied") {
		t.Fatalf("expected root cause in chain, got %q", errorValue)
	}

	cause, _ := fields["error_cause"].(string)
	if !strings.Contains(cause, "AccessDenied") {
		t.Fatalf("expected AccessDenied in error_cause, got %q", cause)
	}
	if got := err.Error(); got != "Wasabi upload failed" {
		t.Fatalf("Error() must stay client-safe, got %q", got)
	}
}

func TestLogFieldsForPlainError(t *testing.T) {
	fields := LogFields(errors.New("db timeout"))
	if fields["error"] != "db timeout" {
		t.Fatalf("unexpected error field: %#v", fields["error"])
	}
	if _, ok := fields["error_cause"]; ok {
		t.Fatalf("did not expect error_cause for plain error, got %#v", fields)
	}
}

func TestFilterStackFramesRequiresConfiguredPrefixes(t *testing.T) {
	frames := []StackFrame{
		{File: "internal/presentation/http/handlers/internalapi/logging_scenario_steps.go", Line: 18, Function: "queryLoggingScenario"},
	}

	if got := FilterStackFrames(frames, nil); got != nil {
		t.Fatalf("expected nil without include prefixes, got %#v", got)
	}

	filtered := FilterStackFrames(frames, []string{"internal/"})
	if len(filtered) != 1 {
		t.Fatalf("expected one filtered frame, got %#v", filtered)
	}
}

func TestStripStackFramePrefixesMakesPathsRelative(t *testing.T) {
	frames := []StackFrame{
		{
			File:     "/tmp/example-service/internal/presentation/http/handlers/internalapi/logging_scenario_steps.go",
			Line:     18,
			Function: "queryLoggingScenario",
		},
	}

	trimmed := StripStackFramePrefixes(frames, []string{"/tmp/example-service"})
	if got := FormatStackFrame(trimmed[0]); got != "internal/presentation/http/handlers/internalapi/logging_scenario_steps.go:18 queryLoggingScenario" {
		t.Fatalf("unexpected trimmed frame: %q", got)
	}
}

func TestStackFramesProvideChainDerivedFramesForWrappedError(t *testing.T) {
	base := Internal("logging scenario failed", errors.New("logging scenario database timeout"), nil)
	level1 := Wrap(base, "failed to execute logging scenario")
	level2 := Wrap(level1, "failed to load logging scenario")

	frames := StackFrames(level2)
	if len(frames) == 0 {
		t.Fatalf("expected stack frames")
	}
	if !strings.Contains(FormatStackFrame(frames[0]), "error_test.go") {
		t.Fatalf("expected caller-derived frame, got %#v", frames)
	}
}

// error frames should include the location even if error is just returned as is and not wrapped
// so in this test all functions returning error should be seen in frames
func TestStackFramesForFunc1Func2Func3Func4Scenario(t *testing.T) {
	err := scenarioFunc4InspectError()
	if err == nil {
		t.Fatalf("expected scenario error")
	}

	frames := StackFrames(err)
	if len(frames) == 0 {
		t.Fatalf("expected stack frames")
	}

	formatted := make([]string, 0, len(frames))
	for _, frame := range frames {
		formatted = append(formatted, FormatStackFrame(frame))
	}
	joined := strings.Join(formatted, "\n")

	if !strings.Contains(joined, "scenarioFunc1Fail") {
		t.Fatalf("expected func1 in stack frames, got %q", joined)
	}
	if !strings.Contains(joined, "scenarioFunc2Wrap") {
		t.Fatalf("expected func2 in stack frames, got %q", joined)
	}
	if !strings.Contains(joined, "scenarioFunc3Return") {
		t.Fatalf("did not expect func3 in stack when only returning err, got %q", joined)
	}
}

func scenarioFunc1Fail() error {
	return cerrors.New("something happened")
}

func scenarioFunc2Wrap() error {
	err := scenarioFunc1Fail()
	if err != nil {
		return Internal("err calling func1", err, map[string]any{
			"func1": "scenarioFunc1Fail",
		})
	}
	return nil
}

func scenarioFunc3Return() error {
	err := scenarioFunc2Wrap()
	if err != nil {
		return err
	}
	return nil
}

func scenarioFunc4InspectError() error {
	err := scenarioFunc3Return()
	if err != nil {
		return err
	}
	return nil
}

func TestFluentErrorUsage(t *testing.T) {
	gotErr := savePaymentForTest("pay_123")

	gotFluent, ok := gotErr.(*fluentError)
	if !ok {
		t.Fatalf("expected *fluentError, got %T", gotErr)
	}
	got := gotFluent.value
	if got.Code != CodeInternal {
		t.Fatalf("expected code %s, got %s", CodeInternal, got.Code)
	}
	if got.Message != "failed to persist payment" {
		t.Fatalf("unexpected message: %q", got.Message)
	}
	if got.Err == nil {
		t.Fatalf("expected wrapped source error")
	}
	if got.Fields["paymentId"] != "pay_123" {
		t.Fatalf("expected paymentId field to be set")
	}
}

func savePaymentForTest(paymentID string) error {
	if err := io.EOF; err != nil {
		return InternalError().
			Message("failed to persist payment").
			Err(err).
			Fields(map[string]any{"paymentId": paymentID})
	}
	return nil
}
