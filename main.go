/*

This package attempts to solve the problem of detecting cycles in a graph
in an embarassingly parallel way: computing the nth power of the adjacency
matrix of the graph, and checking that the trace of this matrix is zero[1],
where n is the dimension of the matrix.

To do this, I am creating a data pipeline:

   1. The initial input will be the original adjacency matrix.
   2. Each row and column will be parsed, and passed in order to a single
      channel. Wrapper types are used to preserve row number and column
      number respectively.
   3. A series of goroutines consume from this channel, computing the resulting
      products and placing them in a final channel. A watcher process consumes
      from this channel, painstakingly pieces each element into the new matrix,
      and then sends it back.
   4. The process repeats until all powers have been computed.

[1] This observation relies on the fact that the elements of the nth power
of the adjacency matrix represent the number of n-length paths between each
node. An acyclic graph will not have any possible paths from a node to itself,
so the diagonal elements should all be zero. Why do we need to compute the nth
power and not just the first power? Imagine a graph arrayed circularly, and you'll
have your answer.

*/

package main

import (
    "log"
    "encoding/csv"
	"./matrix"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
	"strconv"
)

// This function takes in an adjacency matrix (properties: must be square), and returns
// whether or not it contains a cycle. This will work for both directed and undirected
// graphs.
func isGraphCyclic(adjacencyMatrix matrix.Matrix) bool {

	// raise the adjacency matrix to A^len(A.Rows). 
	raisedAdjacencyMatrix := adjacencyMatrix.Exponentiate(adjacencyMatrix.Rows)

	// compute the trace and check if it is not zero, in which case it is cyclic
	return raisedAdjacencyMatrix.Trace() != 0
}

// return a slice of slice of ints from a slice of slice of strings.
func convertContentsToInt(contents [][]string) [][]int {

    newContents := make([][]int, len(contents))

	for lineIndex, line := range contents {

		newContents[lineIndex] = make([]int, len(line))

		for entryIndex, entry := range line {

			intifiedEntry, err := strconv.Atoi(entry)

			if (err != nil) {
				log.Fatal(err)
				panic(err)
			}

			newContents[lineIndex][entryIndex] = intifiedEntry

		}
	}

	return newContents
}

func main() {

	filename := flag.String("file-location", "", "Path to a CSV file containing our desired matrix")
	flag.Parse()

	if *filename != "" {

		// read the file
		fileContents, err := ioutil.ReadFile(*filename)
		if err != nil {
			log.Fatal(err)
		}

		// parse the file
		reader := csv.NewReader(strings.NewReader(string(fileContents)))
		elements, err := reader.ReadAll()
		if err != nil {
			log.Fatal(err)
			panic(err)
		}

		// elements is of type [][]string, but we need [][]int
		matrixElements := convertContentsToInt(elements)
		matrixRows, matrixColumns := len(matrixElements), len(matrixElements)
		if matrixRows != matrixColumns {
			fmt.Println("WARN: Only square matrices allowed.")
		}

		constructedMatrix, err := matrix.NewMatrix(matrixRows, matrixColumns, matrixElements)

		if err != nil {
			log.Fatal(err)
			panic(err)
		}

		fmt.Println("Original matrix:")
		fmt.Println(constructedMatrix)
		fmt.Println("Is it cyclic?")
		fmt.Println(isGraphCyclic(constructedMatrix))

	}

}
