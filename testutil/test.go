package testutil

// Case is a BDD style test case for a feature.
//
// Background: https://en.wikipedia.org/wiki/Behavior-driven_development.
// Aha, maybe you don't need to fully understand it,
// we just use Scenario, Given, When, Then to describe a test case, which has several advantages:
//  1. Highly readable and easy to maintain.
//  2. It can be used as a requirement or use case description, helping you in the TDD process,
//     let AI generate code, that is "code as prompt".
//  3. Test against requirement descriptions, not implementation details.
//     Top-down, and the requirement level is above the details.
//
// To add, "requirement" here is a broad concept, not or not only refers to the requirements
// from the product side, but the interface behavior defined by the test module.
//
// TODO: Use this consistently.
type Case[T1 any, T2 any, T3 any] struct {
	// Scenario describes feature of the test case.
	// E.g. "Query hot block with label filters".
	Scenario string

	// Given is initial "context"!!!(context != parameters of method)
	// at the beginning of the scenario, in one or more clauses.
	// E.g. "a hot block with indexed feeds".
	Given string
	// When is the event that triggers the scenario.
	// E.g. "querying with label filters".
	When string
	// Then is the expected outcome, in one or more clauses.
	// E.g. "should return matching feeds".
	Then string

	// GivenDetail is the detail of the given context.
	// Generally speaking, it describes what "state the object" of the module should have.
	// E.g. 'hot block', what does it look like, what are its member variable values?
	// What is the expected behavior of external dependencies?
	GivenDetail T1
	// WhenDetail is the detail of the when event.
	// Generally speaking, it describes the "parameters of the method call".
	// E.g. what does the query options look like.
	WhenDetail T2
	// ThenExpected is the expected outcome of the scenario.
	// Generally speaking, it describes the "return value of the method call".
	// E.g. what does the returned feeds look like.
	ThenExpected T3
}
