// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package pool

import (
	"context"
	"fmt"
)

// BatchOptions controls full retry passes for Pool.Batch.
type BatchOptions struct {
	// BonusPasses is the number of passes over requests that failed earlier.
	// Zero performs only the initial pass.
	BonusPasses int
	// Retry decides whether a failed request enters the next pass.
	// A nil function retries every failed request.
	Retry func(error) bool
}

// BatchResult contains the final result for each request completed by Batch.
type BatchResult struct {
	// Results are ordered by their original position in the input slice.
	// Requests that could not be completed before context cancellation are absent.
	Results []Result
	// Recovered is the number of requests that succeeded after an earlier failure.
	Recovered int
}

// batchItem keeps the original request index across shrinking retry passes.
type batchItem struct {
	index   uint64
	request Request
}

// batchPassResult associates one stream result with its original request.
type batchPassResult struct {
	item   batchItem
	result Result
}

// Batch executes requests in full passes and optionally retries failed items.
//
// The first pass contains every request.
// Each bonus pass contains only requests that failed
// in the previous pass and match BatchOptions.Retry.
// The pool timeout is applied independently when each pass submits its requests.
// Use the caller context to impose one total deadline across all passes.
func (p *Pool) Batch(
	ctx context.Context,
	requests []Request,
	options BatchOptions,
) (BatchResult, error) {
	if p == nil {
		return BatchResult{}, ErrClosed
	}
	if ctx == nil {
		return BatchResult{}, ErrNilContext
	}
	if err := p.ensureOpen(); err != nil {
		return BatchResult{}, err
	}
	if options.BonusPasses < 0 {
		return BatchResult{}, fmt.Errorf(
			"%w: bonus passes must not be negative",
			ErrInvalidConfig,
		)
	}
	if len(requests) == 0 {
		return BatchResult{}, nil
	}

	pending := make([]batchItem, len(requests))
	for index, request := range requests {
		pending[index] = batchItem{
			index:   uint64(index),
			request: request,
		}
	}

	final := make(map[uint64]Result, len(requests))
	failedBefore := make(map[uint64]struct{}, len(requests))
	recovered := 0

	for pass := 0; pass <= options.BonusPasses && len(pending) > 0; pass++ {
		passResults, err := p.runBatchPass(ctx, pending)
		next := make([]batchItem, 0, len(passResults))

		for _, passResult := range passResults {
			result := passResult.result
			result.Index = passResult.item.index
			final[passResult.item.index] = result

			if result.Err == nil {
				if _, ok := failedBefore[passResult.item.index]; ok {
					recovered++
				}
				continue
			}

			if pass == options.BonusPasses || !shouldRetryBatchError(options.Retry, result.Err) {
				continue
			}

			failedBefore[passResult.item.index] = struct{}{}
			next = append(next, passResult.item)
		}

		if err != nil {
			return finishBatch(final, len(requests), recovered), err
		}
		if pass == options.BonusPasses {
			break
		}

		pending = next
	}

	return finishBatch(final, len(requests), recovered), nil
}

// runBatchPass executes one bounded stream pass over pending items.
func (p *Pool) runBatchPass(
	ctx context.Context,
	pending []batchItem,
) ([]batchPassResult, error) {
	requests := make(chan Request)
	results := p.Stream(ctx, requests)
	submitted := make(chan struct{})

	go func() {
		defer close(requests)
		defer close(submitted)

		for _, item := range pending {
			select {
			case requests <- item.request:
			case <-ctx.Done():
				return
			case <-p.done:
				return
			}
		}
	}()

	passResults := make([]batchPassResult, 0, len(pending))
	for result := range results {
		if result.Index >= uint64(len(pending)) {
			continue
		}

		passResults = append(passResults, batchPassResult{
			item:   pending[result.Index],
			result: result,
		})
	}
	<-submitted

	if err := ctx.Err(); err != nil {
		return passResults, err
	}
	if err := p.root.Err(); err != nil {
		return passResults, ErrClosed
	}

	return passResults, nil
}

// shouldRetryBatchError applies the optional batch retry predicate.
func shouldRetryBatchError(retry func(error) bool, err error) bool {
	if retry == nil {
		return true
	}

	return retry(err)
}

// finishBatch orders final results by their original input position.
func finishBatch(
	final map[uint64]Result,
	requestCount int,
	recovered int,
) BatchResult {
	results := make([]Result, 0, len(final))
	for index := range requestCount {
		if result, ok := final[uint64(index)]; ok {
			results = append(results, result)
		}
	}

	return BatchResult{
		Results:   results,
		Recovered: recovered,
	}
}
