# About

This package attempts to solve the problem of detecting cycles in a graph
in an embarassingly inefficient way: computing the nth power of the adjacency
matrix of the graph, and checking that the trace of this matrix is zero[1],
where n is the dimension of the matrix. It exists purely so I can get better 
at Go.

To do this, I am creating a data pipeline:

   1. The initial input will be the original adjacency matrix.
   2. Each row and column is parsed, and passed in order to a single
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
