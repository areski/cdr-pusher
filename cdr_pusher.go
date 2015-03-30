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
	log "github.com/Sirupsen/logrus"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

// Wait time for results in goroutine
const WAITTIME = 60

// RunFetcher fetchs non imported CDRs from the local datasource (SQLite)
func RunFetcher(config Config, chanRes chan map[int][]string, chanSync chan bool) {
	f := new(SQLFetcher)
	if config.StorageSource == "sqlite" {
		f.Init(config.DBFile, config.DBTable, config.MaxPushBatch, config.CDRFields, config.DBFlagField)
		for {
			log.Info("RunFetcher waiting on chanSync before fetching")
			<-chanSync
			// Fetch CDRs from SQLite
			err := f.Fetch()
			if err != nil {
				log.Error(err.Error())
			}
			if err == nil && f.results != nil {
				chanRes <- f.results
			}
			// Wait x seconds between each DB fetch | Heartbeat
			log.Info("RunFetcher sleeps for " + strconv.Itoa(config.Heartbeat) + " seconds!")
			time.Sleep(time.Second * time.Duration(config.Heartbeat))
		}
	}
}

// DispatchPush is a dispacher to push the results to the right storage
func DispatchPush(config Config, results map[int][]string) {
	if config.StorageDestination == "postgres" || config.StorageDestination == "both" {
		// Push CDRs to PostgreSQL
		pc := new(PGPusher)
		pc.Init(config.PGDatasourcename, config.CDRFields, config.SwitchIP, config.CDRSourceType, config.TableDestination)
		err := pc.Push(results)
		if err != nil {
			log.Error(err.Error())
		}
	}
	if config.StorageDestination == "riak" || config.StorageDestination == "both" {
		// Push CDRs to Riak
		rc := new(RiakPusher)
		rc.Init(config.RiakConnect, config.CDRFields, config.SwitchIP, config.CDRSourceType, config.RiakBucket)
		err := rc.Push(results)
		if err != nil {
			log.Error(err.Error())
		}
	}
}

// PushResult is goroutine that will push CDRs to storage when receiving from results
// on channel chanRes
func PushResult(config Config, chanRes chan map[int][]string, chanSync chan bool) {
	for {
		log.Debug("PushResult sending chanSync to start Fetching")
		// Send signal to go_fetch to fetch
		chanSync <- true
		// waiting for CDRs on channel
		select {
		case results := <-chanRes:
			// Send results to storage engine
			DispatchPush(config, results)
		case <-time.After(time.Second * WAITTIME):
			log.Debug("Nothing received yet...")
		}
	}
}

// PopulateFakeCDR is provided for tests purpose, it takes care of populating the
// SQlite database with fake CDRs at interval of time.
func PopulateFakeCDR(config Config) error {
	if config.FakeCDR != "yes" {
		return nil
	}
	// Heartbeat time for goPopulateFakeCDRs
	intval_time := 1
	for {
		// Wait x seconds when inserting fake CDRs
		log.Info("goPopulateFakeCDRs sleeps for " + strconv.Itoa(intval_time) + " seconds!")
		time.Sleep(time.Second * time.Duration(intval_time))
		GenerateCDR(config.DBFile, config.FakeAmountCDR)
	}
}

// RunApp is the core function of the service it launchs the different goroutines
// that will fetch and push
func RunApp() (string, error) {
	if err := LoadConfig(prodConf); err != nil {
		log.Error(err.Error())
		return "", err
	}
	if err := ValidateConfig(config); err != nil {
		panic(err)
	}

	chanSync := make(chan bool, 1)
	chanRes := make(chan map[int][]string, 1)

	// Start the coroutines
	go RunFetcher(config, chanRes, chanSync)
	go PushResult(config, chanRes, chanSync)
	go PopulateFakeCDR(config)

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
			log.Warn("Got signal:", killSignal)
			if killSignal == os.Interrupt {
				return "Service was interruped by system signal", nil
			}
			return "Service was killed", nil
		}
	}
}

func main() {
	// Log as JSON instead of the default ASCII formatter.
	// log.SetFormatter(&log.JSONFormatter{})

	// Use the Airbrake hook to report errors that have Error severity or above to
	// an exception tracker. You can create custom hooks, see the Hooks section.
	// log.AddHook(&logrus_airbrake.AirbrakeHook{})

	setlogfile := false
	if setlogfile {
		// backendlog := logging.NewLogBackend(os.Stderr, "", 0)
		f, err := os.OpenFile("cdr-pusher.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			panic(err.Error())
		}
		defer f.Close()
		// Output to stderr instead of stdout, could also be a file.
		log.SetOutput(f)
	} else {
		log.SetOutput(os.Stderr)
	}

	// Only log the warning severity or above.
	// log.SetLevel(log.WarnLevel)
	log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	log.Info("StartTime: " + time.Now().Format("Mon Jan _2 2006 15:04:05"))
	RunApp()
	log.Info("StopTime: " + time.Now().Format("Mon Jan _2 2006 15:04:05"))
}
