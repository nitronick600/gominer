package sia

import (
	"log"
	"time"

	"github.com/robvanmieghem/gominer/clients"
	"github.com/robvanmieghem/gominer/mining"
)

//miningWork is sent to the mining routines and defines what ranges should be searched for a matching nonce
type miningWork struct {
	Header []byte
	Offset int
	Job    interface{}
}

// Miner actually mines :-)
type Miner struct {
	HashRateReports   chan *mining.HashRateReport
	miningWorkChannel chan *miningWork
	//Intensity defines the GlobalItemSize in a human friendly way, the GlobalItemSize = 2^Intensity
	Intensity      int
	GlobalItemSize int
	Client         clients.Client
}

//singleDeviceMiner actually mines on 1 opencl device
type singleDeviceMiner struct {
	MinerID           int
	HashRateReports   chan *mining.HashRateReport
	miningWorkChannel chan *miningWork
	//Intensity defines the GlobalItemSize in a human friendly way, the GlobalItemSize = 2^Intensity
	Intensity      int
	GlobalItemSize int
	Client         clients.HeaderReporter
}

//Mine spawns a seperate miner for each device defined in the CLDevices and feeds it with work
func (m *Miner) Mine() {}

const maxUint32 = int64(^uint32(0))

func (m *Miner) createWork() {
	//Register a function to clear the generated work if a job gets deprecated.
	// It does not matter if we clear too many, it is worse to work on a stale job.
	m.Client.SetDeprecatedJobCall(func() {
		numberOfWorkItemsToRemove := len(m.miningWorkChannel)
		for i := 0; i <= numberOfWorkItemsToRemove; i++ {
			<-m.miningWorkChannel
		}
	})

	m.Client.Start()

	for {
		target, header, deprecationChannel, job, err := m.Client.GetHeaderForWork()

		if err != nil {
			log.Println("ERROR fetching work -", err)
			time.Sleep(1000 * time.Millisecond)
			continue
		}

		//copy target to header
		for i := 0; i < 8; i++ {
			header[i+32] = target[7-i]
		}
		//Fill the workchannel with work
		// Only generate nonces for a 32 bit space (since gpu's are mostly 32 bit)
	nonce32loop:
		for i := int64(0); i*int64(m.GlobalItemSize) < (maxUint32 - int64(m.GlobalItemSize)); i++ {
			//Do not continue mining the 32 bit nonce space if the current job is deprecated
			select {
			case <-deprecationChannel:
				break nonce32loop
			default:
			}

			m.miningWorkChannel <- &miningWork{header, int(i) * m.GlobalItemSize, job}
		}
	}
}
