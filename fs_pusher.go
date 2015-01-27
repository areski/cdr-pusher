package main

//
// Prepare PG Database:
//
// $ createdb testdb
// $ psql testdb
// testdb=#
// CREATE TABLE test
//     (id int, call_uuid text, dst text, callerid_name text, callerid_num text, duration int,
//      data jsonb, created timestamp );
//
// INSERT INTO cdr VALUES ("Outbound Call","123555555","123555555","default","2015-01-14 17:58:01","2015-01-14 17:58:01","2015-01-14 17:58:06",5,5,"NORMAL_CLEARING","2bbe83f7-5111-4b5b-9626-c5154608d4ee","","")
//

import (
	"fmt"
	// "github.com/kr/pretty"
	// "github.com/Sirupsen/logrus"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// var log = logrus.New()

// Wait time for results in goroutine
const WAITTIME = 60

// Fetch CDRs from datasource
func gofetcher(config Config, chan_res chan map[int][]string, chan_sync chan bool) {
	for {
		println("gofetcher")
		// TODO: move chan_sync top of f.Fetch and add a loop
		<-chan_sync
		f := new(SQLFetcher)
		if config.Storage_destination == "sqlite" {
			f.Init(config.Db_file, config.Db_table, config.Max_push_batch, config.Cdr_fields)
			// Fetch CDRs from SQLite
			err := f.Fetch()
			if err != nil {
				println(err.Error())
				panic(err)
			}
		}
		chan_res <- f.results
		// Wait x seconds between each DB fetch | Heartbeat
		fmt.Printf("Sleep for %d seconds!\n", config.Heartbeat)
		time.Sleep(time.Second * time.Duration(config.Heartbeat))
	}
}

// Push CDRs to storage
func gopusher(config Config, chan_res chan map[int][]string, chan_sync chan bool) {
	// log.Debug("gopusher")
	for {
		println("gopusher")
		// Send signal to go_fetch to fetch
		chan_sync <- true
		// waiting for CDRs on channel
		select {
		case results := <-chan_res:
			if config.Storage_destination == "postgres" {
				// Push CDRs to PostgreSQL
				p := new(PGPusher)
				p.Init(config.Pg_datasourcename, config.Cdr_fields, config.Switch_ip, config.Table_destination)
				err := p.Push(results)
				if err != nil {
					println(err.Error())
					panic(err)
				}
			}
		case <-time.After(time.Second * WAITTIME):
			fmt.Println("Nothing received yet...")
		}
	}
}

func run_app() (string, error) {
	// log.Info("run_app")
	LoadConfig(Default_conf)
	if err := ValidateConfig(config); err != nil {
		panic(err)
	}

	chan_sync := make(chan bool, 1)
	chan_res := make(chan map[int][]string, 1)

	// Start coroutines
	println("Start coroutines")
	go gofetcher(config, chan_res, chan_sync)
	go gopusher(config, chan_res, chan_sync)

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	// loop work cycle which listen for command or interrupt
	// by system signal
	for {
		select {
		case killSignal := <-interrupt:
			log.Println("Got signal:", killSignal)
			if killSignal == os.Interrupt {
				return "Service was interruped by system signal", nil
			}
			return "Service was killed", nil
		}
	}
	return "", nil
}

func init() {
	// log.Out = os.Stderr
	// // Log as JSON instead of the default ASCII formatter.
	// log.SetFormatter(&log.JSONFormatter{})

	// // Use the Airbrake hook to report errors that have Error severity or above to
	// // an exception tracker. You can create custom hooks, see the Hooks section.
	// // log.AddHook(&logrus_airbrake.AirbrakeHook{})

	// // backendlog := logging.NewLogBackend(os.Stderr, "", 0)
	// f, err := os.OpenFile("testlogfile.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// if err != nil {
	// 	panic(err.Error())
	// }
	// defer f.Close()
	// // Output to stderr instead of stdout, could also be a file.
	// // log.SetOutput(f)
	// log.SetOutput(os.Stderr)

	// // Only log the warning severity or above.
	// // log.SetLevel(log.WarnLevel)
	// log.SetLevel(log.DebugLevel)
}

func main() {

	// log.Debug("debug")
	// log.Info("info")
	// log.Warn("warning")
	// log.Error("err")
	// log.Fatal("fatal")

	// fmt.Printf("StartTime: %v\n", time.Now())
	run_app()
	// fmt.Printf("StopTime: %v\n", time.Now())
}
