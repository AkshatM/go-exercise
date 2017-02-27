package matrix

import "errors"

type Element struct {
	RowIndex int
	ColIndex int
	Value    int
}

type Matrix struct {
	Rows    int
	Columns int
	Entries [][]int
}

// A constructor function for the exported matrix type that offers entry initialisation as a bonus.
// params: rows: The number of rows in the matrix. Must be > 0.
// params: columns: The number of columns in the matrix. Must be > 0/
// params: entries: An actual initializer slice within a slice that will populate the matrix for you.
//         Note that, if provided, the matrix will ignore this if it detects an inconsistency between
//         the provided `rows` or the provided `columns` and the number of rows/columns in `entries`.
func NewMatrix(rows int, columns int, entries ...[][]int) (Matrix, error) {

	if rows <= 0 || columns <= 0 {
		return Matrix{}, errors.New("Both rows and columns must be greater than 0")
	}

	if len(entries) == 0 {
		m := Matrix{rows, columns, make([][]int, rows)}
		for i, _ := range m.Entries {
			m.Entries[i] = make([]int, columns)
		}
		return m, nil
	}

	if len(entries[0]) != rows || len(entries[0][0]) != columns {
		return Matrix{}, errors.New("Provided rows and columns don't match with provided entries")
	}

	return Matrix{rows, columns, entries[0]}, nil
}

// A function that lets you compute integer positive non-zero powers of a matrix.
func (initialMatrix Matrix) Exponentiate(power int) Matrix {

	if power <= 0 {
		panic(errors.New("Only integer positive non-zero powers are allowed"))
	}

	currentMatrix := initialMatrix

	// multiply matrix by itself in parallel
	for i := 1; i < power; i++ {
		currentMatrix = initialMatrix.Multiply(currentMatrix)
	}

	return currentMatrix
}

// A function to left-multiply two matrices together. Panics if the dimensions of the provided matrices
// are not sufficient to provide a final value. Internally, it makes use of goroutines to obtain
// asynchronous multiplication - copies of matrix elements are sent to two input channels, are multiplied
// together by a reader, and piped to a final output channel which then pieces together the final matrix.
// This isn't an improvement over just multiplying them serially, but it serves the purpose of allowing me
// to experiment with goroutines. In the future, I may reiterate on this to implement true parallel multiplication
// - recursively multiply block component submatrices of our chosen matrix.

func (self Matrix) Multiply(secondMatrix Matrix) Matrix {

	if self.Rows != secondMatrix.Columns {
		panic(errors.New("Matrices are not compatible for matrix multiplication - check their dimensions."))
	}

	// declare the shape of our final result
	computedMatrix, err := NewMatrix(self.Rows, self.Columns)

	if err != nil {
		panic(err)
	}

	// all channels need to be buffered in order to avoid deadlock. We pick the total number of elements we'll ever send.
	// If unbuffered, deadlock arises because in `products`, we wait for someone to be listening to the final output channel
	// while program is still trying to send to the first channel before it starts listening to final output, causing all
	// goroutines to get stuck.

	size := computedMatrix.Rows * computedMatrix.Rows * computedMatrix.Rows // Note: this also supports non-square matrices
	firstInputChannel, secondInputChannel := make(chan Element, size), make(chan Element, size)
	finalOutputChannel := make(chan Element, size)

	// start listening for individual elements to process - this goroutine will be blocked until we start sending data below.
	go computeProducts(firstInputChannel, secondInputChannel, finalOutputChannel)

	// Iterate through every element A_ij in the first matrix. For every ij, send all elements of the jth row to the second
	// input channel and len(A[i]) copies of A_ij to the first input channel. This way, we can compute only the pairs we really
	// need to obtain. The order we flush to these channels ensures we get back our data consistently.

	for i, _ := range self.Entries {
		for j, firstValue := range self.Entries[i] {
			for k, secondValue := range secondMatrix.Entries[j] {
				firstInputChannel <- Element{i, j, firstValue}
				secondInputChannel <- Element{j, k, secondValue}
			}
		}
	}

	// close both channels so that our goroutines can exit once needed.
	close(firstInputChannel)
	close(secondInputChannel)

	// read in the final computed values and just build our final computed matrix.
	for computedElement := range finalOutputChannel {
		computedMatrix.Entries[computedElement.RowIndex][computedElement.ColIndex] += computedElement.Value
	}

	return computedMatrix

}

// A go routine that simply reads from two channels, computes products and then send the result elsewhere. Only used
// internally.
func computeProducts(firstInputChannel <-chan Element, secondInputChannel <-chan Element, finalOutputChannel chan<- Element) {

	for {

		x, firstChannelIsOpen := <-firstInputChannel
		y, secondChannelIsOpen := <-secondInputChannel

		if firstChannelIsOpen && secondChannelIsOpen {
			// compute individual product and tag the matrix element the product should belong to.
			finalOutputChannel <- Element{RowIndex: x.RowIndex, ColIndex: y.ColIndex, Value: x.Value * y.Value}

		} else {

			close(finalOutputChannel)
			break

		}
	}
}

// Computes the trace of our matrix type
func (self Matrix) Trace() int {

	trace := 0

	// iterate through the entries and sum the diagonal values
	for i, _ := range self.Entries {
		for j, value := range self.Entries[i] {
			if i == j {
				trace += value
			}
		}
	}

	return trace
}


