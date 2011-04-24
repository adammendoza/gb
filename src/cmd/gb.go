package main

import (
	"log"
	"time"
	"flag"
	"os"
	"github.com/paulosuzart/gb/gbclient"
	"strings"
)

var (
	concurrent = flag.Int("c", 1, "Number of concurrent users emulated. Default 1.")
	requests   = flag.Int("n", 1, "Number of total request to be performed. Default 1.")
	target     = flag.String("t", "http://localhost:8089", "Target to perform the workload.")
	unamePass  = flag.String("A", "", "auth-name:password")
	uname      = ""
	passwd     = ""
)


//Starts concurrent number of workers and waits for everyone terminate. 
//Computes the average time and log it.
func main() {
	flag.Parse()
	log.Print("Starting requests...")

	authData := strings.Split(*unamePass, ":", 1)
	if len(authData) == 2 {
		uname = authData[0]
		passwd = authData[1]
	}

	master := &Master{
		monitor:  make(chan *workSumary),
		ctrlChan: make(chan bool),
	}
	master.BenchMark()

	//wait for the workers to complete after sumarize
	<-master.ctrlChan
	log.Printf("Job done.")

}

//Represents this master.
type Master struct {
	monitor  chan *workSumary
	ctrlChan chan bool
	workers  map[*Worker]Worker
}

//For each client passed by arg, a new worker is created.
//Workers pointers are stored in m.workers to check the end of
//work for each one.
func (m *Master) BenchMark() {
	// starts the sumarize reoutine.
	go m.Sumarize()
	m.workers = map[*Worker]Worker{}

	for c := 0; c < *concurrent; c++ {

		//create a new Worker	
		var w Worker
		w.httpClient = gbclient.NewHTTPClient(*target, "GET")
		w.httpClient.Auth(uname, passwd)
		w.resultChan = m.monitor
		w.work = perform
		w.requests = *requests

		m.workers[&w] = w

		// a go for the Worker
		go w.Execute()
		// #TODO if a worker get stuck it will never send back the result
		// we need a timout for every worker.
	}

}

//Read back the workSumary of each worker.
//Calculates the average response time and total time for the
//whole request.
func (m *Master) Sumarize() {
	var start, end int64
	start = time.Nanoseconds()
	var avg float64 = 0
	totalSuc := 0
	totalErr := 0

	for result := range m.monitor {
		//remove the worker from master
		m.workers[result.Worker] = m.workers[result.Worker], false

		avg = (result.avg + avg) / 2
		totalSuc += result.sucCount
		totalErr += result.errCount

		//if workers still working
		if len(m.workers) == 0 {
			end = time.Nanoseconds()
			break
		}

	}

	log.Printf("Total Go Benchmark time %v miliseconds.", (end-start)/1000000)
	log.Printf("%v requests performed. Average response time %v miliseconds.", totalSuc, avg)
	log.Printf("%v requests lost.", totalErr)
	m.ctrlChan <- true

}

//Reported by the worker through resultChan
type workSumary struct {
	errCount int
	sucCount int
	avg      float64
	Worker   *Worker
}

//A worker
type Worker struct {
	work       func(*gbclient.HTTPClient) (float64, os.Error)
	resultChan chan *workSumary
	httpClient *gbclient.HTTPClient
	requests   int
}

// put the avg response time for the executor.
func (w *Worker) Execute() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			log.Print("Worker died")
		}
	}()

	var totalElapsed float64
	totalErr := 0
	totalSuc := 0

	for i := 0; i < w.requests; i++ {
		elapsed, err := w.work(w.httpClient)
		if err == nil {
			totalSuc += 1
			totalElapsed += elapsed
		} else {
			totalErr += 1
		}
	}

	var sumary workSumary
	sumary.errCount = totalErr
	sumary.sucCount = totalSuc
	sumary.avg = totalElapsed / float64(totalSuc)
	sumary.Worker = w

	w.resultChan <- &sumary

}

func perform(c *gbclient.HTTPClient) (r float64, err os.Error) {
	start := time.Nanoseconds()

	_, err = c.DoRequest()

	if err != nil {
		log.Println(err.String())
		return 0, err
	}

	end := time.Nanoseconds()
	r = float64((end - start) / 1000000)

	return
}
