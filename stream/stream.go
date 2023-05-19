// Copyright 2023 dudaodong@gmail.com. All rights resulterved.
// Use of this source code is governed by MIT license

// Package stream implements a sequence of elements supporting sequential and operations.
// this package is an experiment to explore if stream in go can work as the way java does. its function is very limited.
package stream

import (
	"bytes"
	"encoding/gob"

	"github.com/duke-git/lancet/v2/slice"
	"golang.org/x/exp/constraints"
)

// A stream should implements methods:
// type StreamI[T any] interface {

// 	// part methods of Java Stream Specification.
// 	Distinct() StreamI[T]
// 	Filter(predicate func(item T) bool) StreamI[T]
// 	FlatMap(mapper func(item T) StreamI[T]) StreamI[T]
// 	Map(mapper func(item T) T) StreamI[T]
// 	Peek(consumer func(item T)) StreamI[T]

// 	Sorted(less func(a, b T) bool) StreamI[T]
// 	Max(less func(a, b T) bool) (T, bool)
// 	Min(less func(a, b T) bool) (T, bool)

// 	Limit(maxSize int) StreamI[T]
// 	Skip(n int) StreamI[T]

// 	AllMatch(predicate func(item T) bool) bool
// 	AnyMatch(predicate func(item T) bool) bool
// 	NoneMatch(predicate func(item T) bool) bool
// 	ForEach(consumer func(item T))
// 	Reduce(init T, accumulator func(a, b T) T) T
// 	Count() int

// 	FindFirst() (T, bool)

// 	ToSlice() []T

// 	// part of methods custom extension
// 	Reverse() StreamI[T]
// 	Range(start, end int) StreamI[T]
// 	Concat(streams ...StreamI[T]) StreamI[T]
// }

type stream[T any] struct {
	source []T
}

// Of creates a stream whose elements are the specified values.
// Play: https://go.dev/play/p/jI6_iZZuVFE
func Of[T any](elems ...T) stream[T] {
	return FromSlice(elems)
}

// Generate stream where each element is generated by the provided generater function
// Play: https://go.dev/play/p/rkOWL1yA3j9
func Generate[T any](generator func() func() (item T, ok bool)) stream[T] {
	source := make([]T, 0)

	var zeroValue T
	for next, item, ok := generator(), zeroValue, true; ok; {

		item, ok = next()
		if ok {
			source = append(source, item)
		}
	}

	return FromSlice(source)
}

// FromSlice creates stream from slice.
// Play: https://go.dev/play/p/wywTO0XZtI4
func FromSlice[T any](source []T) stream[T] {
	return stream[T]{source: source}
}

// FromChannel creates stream from channel.
// Play: https://go.dev/play/p/9TZYugGMhXZ
func FromChannel[T any](source <-chan T) stream[T] {
	s := make([]T, 0)

	for v := range source {
		s = append(s, v)
	}

	return FromSlice(s)
}

// FromRange creates a number stream from start to end. both start and end are included. [start, end]
// Play: https://go.dev/play/p/9Ex1-zcg-B-
func FromRange[T constraints.Integer | constraints.Float](start, end, step T) stream[T] {
	if end < start {
		panic("stream.FromRange: param start should be before param end")
	} else if step <= 0 {
		panic("stream.FromRange: param step should be positive")
	}

	l := int((end-start)/step) + 1
	source := make([]T, l)

	for i := 0; i < l; i++ {
		source[i] = start + (T(i) * step)
	}

	return FromSlice(source)
}

// Concat creates a lazily concatenated stream whose elements are all the elements of the first stream followed by all the elements of the second stream.
// Play: https://go.dev/play/p/HM4OlYk_OUC
func Concat[T any](a, b stream[T]) stream[T] {
	source := make([]T, 0)

	source = append(source, a.source...)
	source = append(source, b.source...)

	return FromSlice(source)
}

// Distinct returns a stream that removes the duplicated items.
// Play: https://go.dev/play/p/eGkOSrm64cB
func (s stream[T]) Distinct() stream[T] {
	source := make([]T, 0)

	distinct := map[string]bool{}

	for _, v := range s.source {
		// todo: performance issue
		k := hashKey(v)
		if _, ok := distinct[k]; !ok {
			distinct[k] = true
			source = append(source, v)
		}
	}

	return FromSlice(source)
}

func hashKey(data any) string {
	buffer := bytes.NewBuffer(nil)
	encoder := gob.NewEncoder(buffer)
	err := encoder.Encode(data)
	if err != nil {
		panic("stream.hashKey: get hashkey failed")
	}
	return buffer.String()
}

// Filter returns a stream consisting of the elements of this stream that match the given predicate.
// Play: https://go.dev/play/p/MFlSANo-buc
func (s stream[T]) Filter(predicate func(item T) bool) stream[T] {
	source := make([]T, 0)

	for _, v := range s.source {
		if predicate(v) {
			source = append(source, v)
		}
	}

	return FromSlice(source)
}

// Map returns a stream consisting of the elements of this stream that apply the given function to elements of stream.
// Play: https://go.dev/play/p/OtNQUImdYko
func (s stream[T]) Map(mapper func(item T) T) stream[T] {
	source := make([]T, s.Count())

	for i, v := range s.source {
		source[i] = mapper(v)
	}

	return FromSlice(source)
}

// Peek returns a stream consisting of the elements of this stream, additionally performing the provided action on each element as elements are consumed from the resulting stream.
// Play: https://go.dev/play/p/u1VNzHs6cb2
func (s stream[T]) Peek(consumer func(item T)) stream[T] {
	for _, v := range s.source {
		consumer(v)
	}

	return s
}

// Skip returns a stream consisting of the remaining elements of this stream after discarding the first n elements of the stream.
// If this stream contains fewer than n elements then an empty stream will be returned.
// Play: https://go.dev/play/p/fNdHbqjahum
func (s stream[T]) Skip(n int) stream[T] {
	if n <= 0 {
		return s
	}

	source := make([]T, 0)
	l := len(s.source)

	if n > l {
		return FromSlice(source)
	}

	for i := n; i < l; i++ {
		source = append(source, s.source[i])
	}

	return FromSlice(source)
}

// Limit returns a stream consisting of the elements of this stream, truncated to be no longer than maxSize in length.
// Play: https://go.dev/play/p/qsO4aniDcGf
func (s stream[T]) Limit(maxSize int) stream[T] {
	if s.source == nil {
		return s
	}

	if maxSize < 0 {
		return FromSlice([]T{})
	}

	source := make([]T, 0, maxSize)

	for i := 0; i < len(s.source) && i < maxSize; i++ {
		source = append(source, s.source[i])
	}

	return FromSlice(source)
}

// AllMatch returns whether all elements of this stream match the provided predicate.
// Play: https://go.dev/play/p/V5TBpVRs-Cx
func (s stream[T]) AllMatch(predicate func(item T) bool) bool {
	for _, v := range s.source {
		if !predicate(v) {
			return false
		}
	}

	return true
}

// AnyMatch returns whether any elements of this stream match the provided predicate.
// Play: https://go.dev/play/p/PTCnWn4OxSn
func (s stream[T]) AnyMatch(predicate func(item T) bool) bool {
	for _, v := range s.source {
		if predicate(v) {
			return true
		}
	}

	return false
}

// NoneMatch returns whether no elements of this stream match the provided predicate.
// Play: https://go.dev/play/p/iWS64pL1oo3
func (s stream[T]) NoneMatch(predicate func(item T) bool) bool {
	return !s.AnyMatch(predicate)
}

// ForEach performs an action for each element of this stream.
// Play: https://go.dev/play/p/Dsm0fPqcidk
func (s stream[T]) ForEach(action func(item T)) {
	for _, v := range s.source {
		action(v)
	}
}

// Reduce performs a reduction on the elements of this stream, using an associative accumulation function, and returns an Optional describing the reduced value, if any.
// Play: https://go.dev/play/p/6uzZjq_DJLU
func (s stream[T]) Reduce(initial T, accumulator func(a, b T) T) T {
	for _, v := range s.source {
		initial = accumulator(initial, v)
	}

	return initial
}

// Count returns the count of elements in the stream.
// Play: https://go.dev/play/p/r3koY6y_Xo-
func (s stream[T]) Count() int {
	return len(s.source)
}

// FindFirst returns the first element of this stream and true, or zero value and false if the stream is empty.
// Play: https://go.dev/play/p/9xEf0-6C1e3
func (s stream[T]) FindFirst() (T, bool) {
	var result T

	if s.source == nil || len(s.source) == 0 {
		return result, false
	}

	return s.source[0], true
}

// FindLast returns the last element of this stream and true, or zero value and false if the stream is empty.
// Play: https://go.dev/play/p/WZD2rDAW-2h
func (s stream[T]) FindLast() (T, bool) {
	var result T

	if s.source == nil || len(s.source) == 0 {
		return result, false
	}

	return s.source[len(s.source)-1], true
}

// Reverse returns a stream whose elements are reverse order of given stream.
// Play: https://go.dev/play/p/A8_zkJnLHm4
func (s stream[T]) Reverse() stream[T] {
	l := len(s.source)
	source := make([]T, l)

	for i := 0; i < l; i++ {
		source[i] = s.source[l-1-i]
	}
	return FromSlice(source)
}

// Range returns a stream whose elements are in the range from start(included) to end(excluded) original stream.
// Play: https://go.dev/play/p/indZY5V2f4j
func (s stream[T]) Range(start, end int) stream[T] {
	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = 0
	}
	if start >= end {
		return FromSlice([]T{})
	}

	source := make([]T, 0)

	if end > len(s.source) {
		end = len(s.source)
	}

	for i := start; i < end; i++ {
		source = append(source, s.source[i])
	}

	return FromSlice(source)
}

// Sorted returns a stream consisting of the elements of this stream, sorted according to the provided less function.
// Play: https://go.dev/play/p/XXtng5uonFj
func (s stream[T]) Sorted(less func(a, b T) bool) stream[T] {
	source := []T{}
	source = append(source, s.source...)

	slice.SortBy(source, less)

	return FromSlice(source)
}

// Max returns the maximum element of this stream according to the provided less function.
// less: a > b
// Play: https://go.dev/play/p/fm-1KOPtGzn
func (s stream[T]) Max(less func(a, b T) bool) (T, bool) {
	var max T

	if len(s.source) == 0 {
		return max, false
	}

	for i, v := range s.source {
		if less(v, max) || i == 0 {
			max = v
		}
	}
	return max, true
}

// Min returns the minimum element of this stream according to the provided less function.
// less: a < b
// Play: https://go.dev/play/p/vZfIDgGNRe_0
func (s stream[T]) Min(less func(a, b T) bool) (T, bool) {
	var min T

	if len(s.source) == 0 {
		return min, false
	}

	for i, v := range s.source {
		if less(v, min) || i == 0 {
			min = v
		}
	}

	return min, true
}

// ToSlice return the elements in the stream.
// Play: https://go.dev/play/p/jI6_iZZuVFE
func (s stream[T]) ToSlice() []T {
	return s.source
}
