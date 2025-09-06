package arena_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/pavanmanishd/arena"
)

// BenchmarkWebServerScenarios simulates real web server workloads
func BenchmarkWebServerScenarios(b *testing.B) {

	// HTTP request handler simulation
	b.Run("HTTPRequestHandler", func(b *testing.B) {
		b.Run("Arena", func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Each request gets its own arena
				a := arena.NewArena(8192) // 8KB per request

				// Simulate request processing
				requestHeaders := arena.AllocSlice[string](a, 20) // HTTP headers
				requestBody := a.AllocBytes(1024)                 // Request body buffer
				responseBody := a.AllocBytes(2048)                // Response body buffer
				tempObjects := arena.AllocSlice[int64](a, 50)     // Temporary processing data

				// Simulate some work
				for j := range requestHeaders {
					requestHeaders[j] = "header"
				}
				requestBody[0] = 1
				responseBody[0] = 2
				tempObjects[0] = 3

				// Request complete - arena automatically cleaned up
				a.Release()
			}
		})

		b.Run("Builtin", func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Simulate request processing with regular allocations
				requestHeaders := make([]string, 20)
				requestBody := make([]byte, 1024)
				responseBody := make([]byte, 2048)
				tempObjects := make([]int64, 50)

				// Simulate some work
				for j := range requestHeaders {
					requestHeaders[j] = "header"
				}
				requestBody[0] = 1
				responseBody[0] = 2
				tempObjects[0] = 3

				// Let GC clean up
			}
		})
	})

	// Connection pool simulation
	b.Run("ConnectionPool", func(b *testing.B) {
		const numConnections = 100

		b.Run("Arena_PerConnection", func(b *testing.B) {
			// Each connection has its own arena
			arenas := make([]*arena.Arena, numConnections)
			for i := range arenas {
				arenas[i] = arena.NewArena(4096)
			}
			defer func() {
				for _, a := range arenas {
					a.Release()
				}
			}()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				connID := i % numConnections
				a := arenas[connID]

				// Simulate connection-specific temporary data
				buffer := a.AllocBytes(256)
				metadata := arena.Alloc[int64](a)

				buffer[0] = byte(i)
				*metadata = int64(i)

				// Reset connection arena periodically
				if i%1000 == 999 {
					a.Reset()
				}
			}
		})

		b.Run("Builtin", func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Simulate connection-specific temporary data
				buffer := make([]byte, 256)
				metadata := new(int64)

				buffer[0] = byte(i)
				*metadata = int64(i)
			}
		})
	})
}

// BenchmarkDatabaseScenarios simulates database operation workloads
func BenchmarkDatabaseScenarios(b *testing.B) {

	type DatabaseRow struct {
		ID        int64
		Name      string
		Email     string
		Data      [128]byte
		CreatedAt time.Time
	}

	b.Run("QueryResultProcessing", func(b *testing.B) {
		const rowsPerQuery = 1000

		b.Run("Arena", func(b *testing.B) {
			a := arena.NewArena(512 * 1024) // 512KB arena
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Simulate processing query results
				rows := arena.AllocSlice[DatabaseRow](a, rowsPerQuery)

				// Populate rows (simulate database driver work)
				for j := range rows {
					rows[j].ID = int64(j)
					rows[j].Name = "John Doe"
					rows[j].Email = "john@example.com"
					rows[j].CreatedAt = time.Now()
				}

				// Process rows (simulate business logic)
				var sum int64
				for _, row := range rows {
					sum += row.ID
				}

				// Reset arena after processing query
				a.Reset()
			}
		})

		b.Run("Builtin", func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Simulate processing query results
				rows := make([]DatabaseRow, rowsPerQuery)

				// Populate rows
				for j := range rows {
					rows[j].ID = int64(j)
					rows[j].Name = "John Doe"
					rows[j].Email = "john@example.com"
					rows[j].CreatedAt = time.Now()
				}

				// Process rows
				var sum int64
				for _, row := range rows {
					sum += row.ID
				}
			}
		})
	})

	b.Run("TransactionProcessing", func(b *testing.B) {
		type Transaction struct {
			ID       int64
			FromID   int64
			ToID     int64
			Amount   float64
			Metadata map[string]string
		}

		b.Run("Arena", func(b *testing.B) {
			a := arena.NewArena(64 * 1024)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Process a batch of transactions
				transactions := arena.AllocSlice[Transaction](a, 100)

				for j := range transactions {
					transactions[j].ID = int64(j)
					transactions[j].FromID = int64(j * 2)
					transactions[j].ToID = int64(j*2 + 1)
					transactions[j].Amount = float64(j * 100)
					transactions[j].Metadata = make(map[string]string)
					transactions[j].Metadata["type"] = "transfer"
				}

				// Validate and process transactions
				for _, tx := range transactions {
					if tx.Amount > 0 {
						// Simulate processing
						_ = tx.FromID + tx.ToID
					}
				}

				a.Reset()
			}
		})

		b.Run("Builtin", func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Process a batch of transactions
				transactions := make([]Transaction, 100)

				for j := range transactions {
					transactions[j].ID = int64(j)
					transactions[j].FromID = int64(j * 2)
					transactions[j].ToID = int64(j*2 + 1)
					transactions[j].Amount = float64(j * 100)
					transactions[j].Metadata = make(map[string]string)
					transactions[j].Metadata["type"] = "transfer"
				}

				// Validate and process transactions
				for _, tx := range transactions {
					if tx.Amount > 0 {
						// Simulate processing
						_ = tx.FromID + tx.ToID
					}
				}
			}
		})
	})
}

// BenchmarkJSONProcessingScenarios simulates JSON parsing/serialization workloads
func BenchmarkJSONProcessingScenarios(b *testing.B) {

	type JSONObject struct {
		ID       int64
		Name     string
		Value    float64
		Tags     []string
		Children []*JSONObject
	}

	b.Run("JSONDocumentParsing", func(b *testing.B) {
		b.Run("Arena", func(b *testing.B) {
			a := arena.NewArena(256 * 1024)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Simulate parsing a complex JSON document
				root := arena.Alloc[JSONObject](a)
				root.ID = int64(i)
				root.Name = "root"
				root.Value = 3.14159
				root.Tags = arena.AllocSlice[string](a, 5)
				root.Children = arena.AllocSlice[*JSONObject](a, 10)

				// Create child objects
				for j := range root.Children {
					child := arena.Alloc[JSONObject](a)
					child.ID = int64(j)
					child.Name = fmt.Sprintf("child_%d", j)
					child.Value = float64(j) * 2.5
					child.Tags = arena.AllocSlice[string](a, 3)

					for k := range child.Tags {
						child.Tags[k] = fmt.Sprintf("tag_%d", k)
					}

					root.Children[j] = child
				}

				// Simulate processing the parsed data
				var sum float64
				for _, child := range root.Children {
					sum += child.Value
				}

				a.Reset()
			}
		})

		b.Run("Builtin", func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Simulate parsing a complex JSON document
				root := &JSONObject{
					ID:    int64(i),
					Name:  "root",
					Value: 3.14159,
					Tags:  make([]string, 5),
				}
				root.Children = make([]*JSONObject, 10)

				// Create child objects
				for j := range root.Children {
					child := &JSONObject{
						ID:    int64(j),
						Name:  fmt.Sprintf("child_%d", j),
						Value: float64(j) * 2.5,
						Tags:  make([]string, 3),
					}

					for k := range child.Tags {
						child.Tags[k] = fmt.Sprintf("tag_%d", k)
					}

					root.Children[j] = child
				}

				// Simulate processing the parsed data
				var sum float64
				for _, child := range root.Children {
					sum += child.Value
				}
			}
		})
	})
}

// BenchmarkGraphAlgorithmScenarios simulates graph processing workloads
func BenchmarkGraphAlgorithmScenarios(b *testing.B) {

	type GraphNode struct {
		ID       int
		Value    int64
		Edges    []*GraphNode
		Visited  bool
		Distance int
		Parent   *GraphNode
	}

	b.Run("GraphTraversal", func(b *testing.B) {
		const numNodes = 1000

		b.Run("Arena", func(b *testing.B) {
			a := arena.NewArena(1024 * 1024) // 1MB arena
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Create graph nodes
				nodes := arena.AllocSlice[*GraphNode](a, numNodes)
				for j := range nodes {
					nodes[j] = arena.Alloc[GraphNode](a)
					nodes[j].ID = j
					nodes[j].Value = int64(j * 2)
					nodes[j].Edges = arena.AllocSlice[*GraphNode](a, 5) // 5 edges per node
				}

				// Connect nodes (create edges)
				for j, node := range nodes {
					for k := range node.Edges {
						targetID := (j + k + 1) % numNodes
						node.Edges[k] = nodes[targetID]
					}
				}

				// Simulate graph traversal (BFS-like)
				queue := arena.AllocSlice[*GraphNode](a, numNodes)
				queueStart, queueEnd := 0, 1
				queue[0] = nodes[0]
				nodes[0].Visited = true
				nodes[0].Distance = 0

				for queueStart < queueEnd {
					current := queue[queueStart]
					queueStart++

					for _, neighbor := range current.Edges {
						if neighbor != nil && !neighbor.Visited {
							neighbor.Visited = true
							neighbor.Distance = current.Distance + 1
							neighbor.Parent = current
							if queueEnd < len(queue) {
								queue[queueEnd] = neighbor
								queueEnd++
							}
						}
					}
				}

				a.Reset()
			}
		})

		b.Run("Builtin", func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Create graph nodes
				nodes := make([]*GraphNode, numNodes)
				for j := range nodes {
					nodes[j] = &GraphNode{
						ID:    j,
						Value: int64(j * 2),
						Edges: make([]*GraphNode, 5),
					}
				}

				// Connect nodes (create edges)
				for j, node := range nodes {
					for k := range node.Edges {
						targetID := (j + k + 1) % numNodes
						node.Edges[k] = nodes[targetID]
					}
				}

				// Simulate graph traversal (BFS-like)
				queue := make([]*GraphNode, numNodes)
				queueStart, queueEnd := 0, 1
				queue[0] = nodes[0]
				nodes[0].Visited = true
				nodes[0].Distance = 0

				for queueStart < queueEnd {
					current := queue[queueStart]
					queueStart++

					for _, neighbor := range current.Edges {
						if neighbor != nil && !neighbor.Visited {
							neighbor.Visited = true
							neighbor.Distance = current.Distance + 1
							neighbor.Parent = current
							if queueEnd < len(queue) {
								queue[queueEnd] = neighbor
								queueEnd++
							}
						}
					}
				}
			}
		})
	})
}

// BenchmarkConcurrentWorkloadScenarios tests concurrent scenarios
func BenchmarkConcurrentWorkloadScenarios(b *testing.B) {

	b.Run("WorkerPoolPattern", func(b *testing.B) {
		const numWorkers = 8
		const jobsPerWorker = 100

		b.Run("Arena_PerWorker", func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				var wg sync.WaitGroup
				wg.Add(numWorkers)

				for w := 0; w < numWorkers; w++ {
					go func(workerID int) {
						defer wg.Done()

						// Each worker gets its own arena
						a := arena.NewArena(64 * 1024)
						defer a.Release()

						for j := 0; j < jobsPerWorker; j++ {
							// Simulate job processing
							buffer := a.AllocBytes(512)
							result := arena.Alloc[int64](a)

							// Do some work
							buffer[0] = byte(workerID)
							*result = int64(workerID*jobsPerWorker + j)

							if j%50 == 49 {
								a.Reset()
							}
						}
					}(w)
				}

				wg.Wait()
			}
		})

		b.Run("SafeArena_Shared", func(b *testing.B) {
			s := arena.NewSafeArena(512 * 1024)
			defer s.Release()

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				var wg sync.WaitGroup
				wg.Add(numWorkers)

				for w := 0; w < numWorkers; w++ {
					go func(workerID int) {
						defer wg.Done()

						for j := 0; j < jobsPerWorker; j++ {
							// Simulate job processing with shared arena
							buffer := s.AllocBytes(512)
							result := arena.SafeAlloc[int64](s)

							// Do some work
							buffer[0] = byte(workerID)
							*result = int64(workerID*jobsPerWorker + j)
						}
					}(w)
				}

				wg.Wait()
				s.Reset()
			}
		})

		b.Run("Builtin", func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				var wg sync.WaitGroup
				wg.Add(numWorkers)

				for w := 0; w < numWorkers; w++ {
					go func(workerID int) {
						defer wg.Done()

						for j := 0; j < jobsPerWorker; j++ {
							// Simulate job processing
							buffer := make([]byte, 512)
							result := new(int64)

							// Do some work
							buffer[0] = byte(workerID)
							*result = int64(workerID*jobsPerWorker + j)
						}
					}(w)
				}

				wg.Wait()
			}
		})
	})
}
