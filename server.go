package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kardianos/service"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

var logger service.Logger

type stats struct {
	Mem float64 `json:"mem"`
	CPU float64 `json:"cpu"`
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	var s stats
	cpu, _ := cpu.Percent(0, false)
	// always single value if percpu is set to false
	s.CPU = cpu[0]
	m, _ := mem.VirtualMemory()
	s.Mem = m.UsedPercent

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

func specsHandler(w http.ResponseWriter, r *http.Request) {
	spec := make(map[string]interface{}, 0)

	spec["cpu_info"], _ = cpu.Info()
	spec["mem_info"], _ = mem.VirtualMemory()
	spec["host_info"], _ = host.Info()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(spec)
}

func addV1Routes(r *mux.Router) {
	r.HandleFunc("/stats", statsHandler).Methods("GET")
	r.HandleFunc("/specs", specsHandler).Methods("GET")
}

type program struct{}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}
func (p *program) run() {
	r := mux.NewRouter().StrictSlash(true)

	addV1Routes(r.PathPrefix("/v1").Subrouter())
	log.Fatal(http.ListenAndServe(":5000", r))
}

func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	return nil
}

func main() {
	svcConfig := &service.Config{
		Name:        "GoPerfMon",
		DisplayName: "EIP Performance Monitor",
		Description: "EngageIP performance monitor written in Golang.",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}
	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}
