package equihash

import (
	"log"
	"reflect"
	"strings"

	"github.com/robvanmieghem/gominer/clients"
	"github.com/robvanmieghem/gominer/clients/stratum"
)

func NewClient(connectionstring, pooluser string, poolPass string) (client clients.Client) {
	if strings.HasPrefix(connectionstring, "stratum+tcp://") {
		client = &EquihashClient{connectionstring: strings.TrimPrefix(connectionstring, "stratum+tcp://"), User: pooluser, Password: poolPass}
	}
	return
}

//Start connects to the stratumserver and processes the notifications
func (ec *EquihashClient) Start() {
	ec.Lock()
	defer func() {
		ec.Unlock()
	}()

	ec.DeprecateOutstandingJobs()

	ec.stratumclient = &stratum.Client{}
	//In case of an error, drop the current stratumclient and restart
	ec.stratumclient.ErrorCallback = func(err error) {
		log.Println("Error in connection to stratumserver:", err)
		ec.stratumclient.Close()
		ec.Start()
	}

	ec.subscribeToStratumDifficultyChanges()
	ec.subscribeToStratumJobNotifications()

	//Connect to the stratum server
	log.Println("Connecting to", ec.connectionstring)
	ec.stratumclient.Dial(ec.connectionstring)

	//Subscribe for mining
	//Close the connection on an error will cause the client to generate an error, resulting in te errorhandler to be triggered
	result, err := ec.stratumclient.Call("mining.subscribe", []string{"cgminer-fuddware/4.9.0"})
	if err != nil {
		log.Println("ERROR Error in response from stratum:", err)
		ec.stratumclient.Close()
		return
	}
	reply, _ := result.([]interface{})
	if !ok || len(reply) < 3 {
		log.Println("ERROR Invalid response from stratum:", result)
		ec.stratumclient.Close()
		return
	}

	//Keep the extranonce1 and extranonce2_size from the reply
	if ec.extranonce1, err = stratum.HexStringToBytes(reply[1]); err != nil {
		log.Println("ERROR Invalid extrannonce1 from startum")
		ec.stratumclient.Close()
		return
	}
	log.Printf("ExtraNonce1: %s", reply[1])
	log.Printf("Subscribe Response: %v", reply)

	extranonce2Size, ok := reply[2].(float64)
	if !ok {
		log.Println("ERROR Invalid extranonce2_size from stratum", reply[2], "type", reflect.TypeOf(reply[2]))
		ec.stratumclient.Close()
		return
	}
	ec.extranonce2Size = uint(extranonce2Size)

	//Authorize the miner
	go func() {
		result, err = sc.stratumclient.Call("mining.authorize", []string{sc.User, sc.Password})
		if err != nil {
			log.Println("Unable to authorize:", err)
			sc.stratumclient.Close()
			return
		}
		log.Println("Authorization of", sc.User, ":", result)
	}()

}

func (sc *EquihashClient) subscribeToStratumDifficultyChanges() {
	sc.stratumclient.SetNotificationHandler("mining.set_difficulty", func(params []interface{}) {
		if params == nil || len(params) < 1 {
			log.Println("ERROR No difficulty parameter supplied by stratum server")
			return
		}
		diff, ok := params[0].(float64)
		if !ok {
			log.Println("ERROR Invalid difficulty supplied by stratum server:", params[0])
			return
		}
		log.Println("Stratum server changed difficulty to", diff)
		sc.setDifficulty(diff)
	})
}

func (sc *EquihashClient) subscribeToStratumJobNotifications() {
	sc.stratumclient.SetNotificationHandler("mining.notify", func(params []interface{}) {
		log.Println("New job received from stratum server")
		// if params == nil || len(params) < 9 {
		// 	log.Println("ERROR Wrong number of parameters supplied by stratum server")
		// 	return
		// }
		log.Printf("Params: %v", params)

		sj := stratumJob{}

		sj.ExtraNonce2.Size = sc.extranonce2Size
		log.Printf("enonce2 %d", sc.extranonce2Size)

		log.Printf("Params: %v", params)
		var ok bool
		var err error
		if sj.JobID, ok = params[0].(string); !ok {
			log.Println("ERROR Wrong job_id parameter supplied by stratum server")
			return
		}
		log.Printf("jobid: %s", sj.JobID)
		if sj.PrevHash, err = stratum.HexStringToBytes(params[1]); err != nil {
			log.Println("ERROR Wrong prevhash parameter supplied by stratum server")
			return
		}
		log.Printf("PrevHash: %s", params[1])
		if sj.Coinbase1, err = stratum.HexStringToBytes(params[2]); err != nil {
			log.Println("ERROR Wrong coinb1 parameter supplied by stratum server")
			return
		}
		log.Printf("Coinbase1: %s", params[2])
		if sj.Coinbase2, err = stratum.HexStringToBytes(params[3]); err != nil {
			log.Println("ERROR Wrong coinb2 parameter supplied by stratum server")
			return
		}
		log.Printf("Coinbase2: %s", params[3])

		//Convert the merklebranch parameter
		merklebranch, ok := params[4].([]interface{})
		if !ok {
			log.Printf("ERROR Wrong merkle_branch parameter supplied by stratum server: %s", params[4])
		}
		log.Printf("merklebranch: %s", merklebranch)

		sj.MerkleBranch = make([][]byte, len(merklebranch), len(merklebranch))
		for i, branch := range merklebranch {
			if sj.MerkleBranch[i], err = stratum.HexStringToBytes(branch); err != nil {
				//log.Printf("ERROR Wrong merkle_branch parameter supplied by stratum server. %s", branch)
				//return
			}
		}

		if sj.Version, ok = params[5].(string); !ok {
			//log.Println("ERROR Wrong version parameter supplied by stratum server")
			//return
		}
		log.Printf("Version: %s", sj.Version)
		if sj.NBits, ok = params[6].(string); !ok {
			//log.Println("ERROR Wrong nbits parameter supplied by stratum server")
			//return
		}
		// log.Printf("NBits: %s", sj.NBits)
		// if sj.NTime, err = stratum.HexStringToBytes(params[7]); err != nil {
		// 	log.Println("ERROR Wrong ntime parameter supplied by stratum server")
		// 	return
		// }
		// log.Printf("NTime: %s", params[7])
		// if sj.CleanJobs, ok = params[8].(bool); !ok {
		// 	log.Println("ERROR Wrong clean_jobs parameter supplied by stratum server")
		// 	return
		// }
		//sc.addNewStratumJob(sj)
	})
}
