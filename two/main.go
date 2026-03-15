package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	// 注册路由
	http.HandleFunc("/health", healthCheck)
	http.HandleFunc("/api/analyze/", analyzeHandler)
	http.HandleFunc("/api/chains/", chainHandler)
	http.HandleFunc("/api/chains", chainsHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server running on port %s\n", port)
	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// 链路结构
type Chain struct {
	ID           string        `json:"id"`
	Upstream     Service       `json:"upstream"`
	Downstreams  []Downstream  `json:"downstreams"`
}

// 服务结构
type Service struct {
	PSM    string `json:"psm"`
	Method string `json:"method"`
}

// 下游服务结构
type Downstream struct {
	Service
	StrongDependency bool `json:"strongDependency"`
}

// 响应结构
type Response struct {
	Status  string      `json:"status,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// 分析结果结构
type AnalysisResult struct {
	Chain       Chain        `json:"chain"`
	StrongDeps  []Downstream `json:"strongDeps"`
	WeakDeps    []Downstream `json:"weakDeps"`
	TotalDeps   int          `json:"totalDeps"`
	StrongCount int          `json:"strongCount"`
	WeakCount   int          `json:"weakCount"`
}

// 模拟数据存储
var chains = []Chain{
	{
		ID: "1",
		Upstream: Service{
			PSM:    "service-a",
			Method: "GET /api/users",
		},
		Downstreams: []Downstream{
			{
				Service: Service{
					PSM:    "service-b",
					Method: "POST /api/auth",
				},
				StrongDependency: true,
			},
			{
				Service: Service{
					PSM:    "service-c",
					Method: "GET /api/config",
				},
				StrongDependency: false,
			},
		},
	},
}

// 健康检查
func healthCheck(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, http.StatusOK, Response{
		Status: "ok",
	})
}

// 处理/chains路由
func chainsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		sendJSON(w, http.StatusOK, chains)
	case "POST":
		createChain(w, r)
	default:
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Error: "Method not allowed",
		})
	}
}

// 处理/chains/{id}路由
func chainHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/chains/")
	if id == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Error: "Missing chain ID",
		})
		return
	}

	switch r.Method {
	case "GET":
		getChain(w, r, id)
	case "PUT":
		updateChain(w, r, id)
	case "DELETE":
		deleteChain(w, r, id)
	default:
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Error: "Method not allowed",
		})
	}
}

// 处理/analyze/{chainId}路由
func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	chainId := strings.TrimPrefix(r.URL.Path, "/api/analyze/")
	if chainId == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Error: "Missing chain ID",
		})
		return
	}

	if r.Method != "GET" {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Error: "Method not allowed",
		})
		return
	}

	analyzeDependencies(w, r, chainId)
}

// 创建链路
func createChain(w http.ResponseWriter, r *http.Request) {
	var chain Chain
	if err := json.NewDecoder(r.Body).Decode(&chain); err != nil {
		sendJSON(w, http.StatusBadRequest, Response{
			Error: err.Error(),
		})
		return
	}

	// 生成ID
	chain.ID = fmt.Sprintf("%d", len(chains)+1)

	chains = append(chains, chain)
	sendJSON(w, http.StatusCreated, chain)
}

// 获取单个链路
func getChain(w http.ResponseWriter, r *http.Request, id string) {
	for _, chain := range chains {
		if chain.ID == id {
			sendJSON(w, http.StatusOK, chain)
			return
		}
	}
	sendJSON(w, http.StatusNotFound, Response{
		Error: "Chain not found",
	})
}

// 更新链路
func updateChain(w http.ResponseWriter, r *http.Request, id string) {
	var updatedChain Chain
	if err := json.NewDecoder(r.Body).Decode(&updatedChain); err != nil {
		sendJSON(w, http.StatusBadRequest, Response{
			Error: err.Error(),
		})
		return
	}

	for i, chain := range chains {
		if chain.ID == id {
			updatedChain.ID = id
			chains[i] = updatedChain
			sendJSON(w, http.StatusOK, updatedChain)
			return
		}
	}
	sendJSON(w, http.StatusNotFound, Response{
		Error: "Chain not found",
	})
}

// 删除链路
func deleteChain(w http.ResponseWriter, r *http.Request, id string) {
	for i, chain := range chains {
		if chain.ID == id {
			chains = append(chains[:i], chains[i+1:]...)
			sendJSON(w, http.StatusOK, Response{
				Message: "Chain deleted",
			})
			return
		}
	}
	sendJSON(w, http.StatusNotFound, Response{
		Error: "Chain not found",
	})
}

// 分析依赖
func analyzeDependencies(w http.ResponseWriter, r *http.Request, chainId string) {
	for _, chain := range chains {
		if chain.ID == chainId {
			// 分析下游依赖
			strongDeps := []Downstream{}
			weakDeps := []Downstream{}

			for _, downstream := range chain.Downstreams {
				if downstream.StrongDependency {
					strongDeps = append(strongDeps, downstream)
				} else {
					weakDeps = append(weakDeps, downstream)
				}
			}

			result := AnalysisResult{
				Chain:       chain,
				StrongDeps:  strongDeps,
				WeakDeps:    weakDeps,
				TotalDeps:   len(chain.Downstreams),
				StrongCount: len(strongDeps),
				WeakCount:   len(weakDeps),
			}

			sendJSON(w, http.StatusOK, result)
			return
		}
	}
	sendJSON(w, http.StatusNotFound, Response{
		Error: "Chain not found",
	})
}

// 发送JSON响应
func sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}
