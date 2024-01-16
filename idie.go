// Demo code for a timer based update
package main

import (
	"flag"
	"fmt"
	"idie/requester"
	"idie/threadman"
	"idie/util"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	updateTimeInterval = 500 * time.Millisecond
)

// ENUM optionOutputType
// do not use iota, to make it more readable
const (
	OUTPUT_TYPE_TXT  = "txt"
	OUTPUT_TYPE_JSON = "json"
	OUTPUT_TYPE_CSV  = "csv"
)

var (
	//appMutex                  sync.Mutex
	resultMutex sync.Mutex
	resultWg    sync.WaitGroup

	app           *tview.Application
	appDesc       *tview.Box
	appDescScreen tcell.Screen
	appFlex       *tview.Flex

	thread     = threadman.NewThreadman(threadman.WithWorkerLimit(optionWorkerLimit))
	resultsMap = make(map[string]map[string][]int)
	totalTask  = 0

	startingTime = time.Now()

	// options & args
	argStartIP        = "" // ipv4 only
	argEndIP          = "" // ipv4 only
	optionPort        = "" // format: 80,443,8080
	optionOutputType  = "" // format: json,txt,csv
	optionOutputFile  = "" // output file path
	optionWorkerLimit = 10 // worker for running task

	// processed options & args
	optionPortProcessed []int
	optionOutputFilePtr *os.File
)

func getThreadStat() string {
	idleCounter := thread.GetStandByCounter()
	runningCounter := thread.GetRunningCounter()
	doneCounter := thread.GetDoneCounter()
	totalCounter := totalTask
	return fmt.Sprintf(" idle:%d running:%d done:%d total:%d ", idleCounter, runningCounter, doneCounter, totalCounter)
}

func elapsedTime() string {
	return fmt.Sprintf(" %s ", time.Since(startingTime).Round(time.Second))
}

// this run with go routine
func updateStat() {
	for {
		time.Sleep(updateTimeInterval)

		if thread.GetDoneCounter() == uint64(totalTask) {
			app.Stop()
			return
		}

		app.QueueUpdateDraw(func() {
			appDesc.Draw(appDescScreen)
		})
	}
}

// this run with go routine
func threadUpdateListener() {
	for {
		select {
		case <-threadman.ThreadInactiveNotifier:
			return
		case task, _ := <-threadman.TaskDoneNotifier:
			go func(tParam *threadman.Task) {
				resultWg.Add(1)
				resultMutex.Lock()
				defer resultMutex.Unlock()
				defer resultWg.Done()

				processTaskDone(tParam)
			}(task)
		}
	}
}

func tcpUdpMapToString(showOpen bool, showClosed bool) (str string) {
	results := [][]string{
		{"IP Address", "Open", "Closed"},
	}

	longestFirstColumn := len("IP Address") // ip column and its value str length
	longestSecondColumn := len("Open")      // open port column and its value str length
	longestThirdColumn := len("Closed")     // closed port column and its value str length

	for ip := range resultsMap {
		if len(ip) > longestFirstColumn {
			longestFirstColumn = len(ip)
		}

		_, tcps, udps, err := breakTcpUdpMap(ip)
		if err != nil {
			panic(err)
		}

		openText := ""
		var openPorts []int
		for i, port := range tcps {
			openPorts = append(openPorts, port)

			openText += fmt.Sprintf("%d", port)
			if i < len(tcps)-1 {
				openText += ","
			}
		}
		for i, port := range udps {
			openPorts = append(openPorts, port)

			openText += fmt.Sprintf("%d", port)
			if i < len(udps)-1 {
				openText += ","
			}
		}

		closedText := ""
		for i, port := range optionPortProcessed {
			if util.IsIntSliceContains(openPorts, port) {
				continue
			}

			closedText += fmt.Sprintf("%d", port)
			if i < len(optionPortProcessed)-1 {
				closedText += ","
			}
		}

		if len(openText) > longestSecondColumn {
			longestSecondColumn = len(openText)
		}
		if len(closedText) > longestThirdColumn {
			longestThirdColumn = len(closedText)
		}

		results = append(results, []string{ip, openText, closedText})
	}

	// add spacing with ' ' rune calculated from (longest column + 2)
	longestFirstColumn += 2
	longestSecondColumn += 2
	longestThirdColumn += 2
	for _, result := range results {
		str += util.FillPostfixWithRune(result[0], longestFirstColumn, ' ')

		if showOpen {
			str += util.FillPostfixWithRune(result[1], longestSecondColumn, ' ')
		}

		if showClosed {
			str += util.FillPostfixWithRune(result[2], longestThirdColumn, ' ')
		}

		str += "\n"
	}

	return
}

func printToFile() {
	str := tcpUdpMapToString(true, true)
	util.WriteStringToFile(optionOutputFilePtr, str)
}

func breakTcpUdpMap(ip string) (tcpUdpMap map[string][]int, tcps []int, udps []int, err error) {
	var ok bool

	tcpUdpMap, ok = resultsMap[ip]
	if !ok {
		err = fmt.Errorf("Error getting tcpUdpMap for ip %s", ip)
		return
	}

	tcps, ok = tcpUdpMap["tcp"]
	if !ok {
		err = fmt.Errorf("Error getting tcps for ip %s", ip)
		return
	}

	udps, ok = tcpUdpMap["udp"]
	if !ok {
		err = fmt.Errorf("Error getting udps for ip %s", ip)
		return
	}

	return
}

func processTaskDone(task *threadman.Task) {
	result, ok := task.Result.(string)
	if !ok {
		return
	}

	// ip,port,tcpBool,udpBool
	resultSplitted := util.Explode(result, ",")
	ip := resultSplitted[0]
	port, _ := strconv.Atoi(resultSplitted[1])
	tcpBool, _ := strconv.ParseBool(resultSplitted[2])
	udpBool, _ := strconv.ParseBool(resultSplitted[3])

	if _, ok = resultsMap[ip]; !ok {
		resultsMap[ip] = make(map[string][]int)
		resultsMap[ip]["tcp"] = []int{}
		resultsMap[ip]["udp"] = []int{}
	}

	_, tcps, udps, err := breakTcpUdpMap(ip)
	if err != nil {
		return
	}

	if tcpBool {
		resultsMap[ip]["tcp"] = append(tcps, port)
		resultsMap[ip]["tcp"] = util.UniqueIntSlice(resultsMap[ip]["tcp"])
	}

	if udpBool {
		resultsMap[ip]["udp"] = append(udps, port)
		resultsMap[ip]["udp"] = util.UniqueIntSlice(resultsMap[ip]["udp"])
	}
}

func wrapperExecutorTask(ip string, port int) interface{} {
	req := requester.NewRequester()
	ipScan, portScan, isOpen, protocol, _ := req.NmapSyn(ip, port)
	isTcpOpen, isUdpOpen := false, false
	if protocol == "tcp" {
		isTcpOpen = isOpen
	}
	if protocol == "udp" {
		isUdpOpen = isOpen
	}
	return interface{}(fmt.Sprintf("%s,%d,%t,%t", ipScan, portScan, isTcpOpen, isUdpOpen))
}

func createDiscovery(start string, end string, ports []int) {
	generatedIPs, err := util.GenerateIPRange(start, end)
	if err != nil {
		return
	}

	for _, ip := range generatedIPs {
		for _, port := range ports {
			lIp := ip
			lPort := port
			thread.AddTask(func() interface{} {
				return wrapperExecutorTask(lIp, lPort)
			})
		}
	}

	totalTask = len(generatedIPs) * len(ports)
}

func drawStatus(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
	appDescScreen = screen

	centerY := y + height/2
	centerY += 2
	progress := float64(thread.GetDoneCounter()) / float64(totalTask)
	progressWidth := int(float64(width) * progress)
	for cx := x + 1; cx < x+progressWidth-1; cx++ {
		screen.SetContent(cx, centerY, '█', nil, tcell.StyleDefault.Foreground(tcell.ColorGreen))
	}

	tview.Print(screen, " v1.0 (by GoogleX) -"+getThreadStat()+"-"+elapsedTime(), x+1, y, width-2, tview.AlignLeft, tcell.ColorWhite)

	// return getinnerrect
	return x + 1, centerY + 3, width - 2, height - (centerY + 3 - y)
}

func prepareFlag() {
	flag.StringVar(&optionPort, "port", "80", "Port to check (format: 80-90,443,25565)")
	flag.StringVar(&optionOutputType, "type", "txt", "Output type (json,txt,csv)")
	flag.StringVar(&optionOutputFile, "file", "", "Output file path")
	flag.IntVar(&optionWorkerLimit, "worker", 10, "Worker limit")
}

func flagValidate() {
	//args
	args := flag.Args()
	if len(args) < 2 {
		fmt.Println("Usage: idie <start ip> <end ip>")
		os.Exit(1)
	}

	argStartIP = args[0]
	if !util.IsValidIPv4(argStartIP) {
		fmt.Println("Invalid start ip")
		os.Exit(1)
	}

	argEndIP = args[1]
	if !util.IsValidIPv4(argEndIP) {
		fmt.Println("Invalid end ip")
		os.Exit(1)
	}

	//option
	if optionOutputFile == "" {
		fmt.Println("Invalid output file (--file)")
		os.Exit(1)
	}

	if optionOutputType != OUTPUT_TYPE_TXT && optionOutputType != OUTPUT_TYPE_JSON && optionOutputType != OUTPUT_TYPE_CSV {
		fmt.Println("Invalid output type (--type)")
		os.Exit(1)
	}

	if optionWorkerLimit <= 0 {
		fmt.Println("Invalid worker limit (--worker)")
		os.Exit(1)
	}

	if !util.IsValidPortList(optionPort) {
		fmt.Println("Invalid port list (--port)")
		os.Exit(1)
	}
}

func threadOptimize() {
	if totalTask < optionWorkerLimit {
		optionWorkerLimit = totalTask / 2
		if optionWorkerLimit <= 0 {
			optionWorkerLimit = 1
		}
	}
	thread.WorkerLimit = optionWorkerLimit
}

func prepareInterface() {
	app = tview.NewApplication()
	appFlex = tview.NewFlex().SetDirection(tview.FlexRow)

	appDesc = tview.NewTextView().
		SetDrawFunc(drawStatus)

	appFlex.AddItem(appDesc, 1, 1, true)

	app = app.SetRoot(appFlex, true).
		SetFocus(appDesc)
}

func main() {
	prepareFlag()
	flag.Parse()
	flagValidate()

	optionOutputFilePtr = util.OpenFileOrCreate(optionOutputFile)

	fmt.Println("Creating task...")
	optionPortProcessed = util.ExplodeToIntSlice(optionPort, ",")
	createDiscovery(argStartIP, argEndIP, optionPortProcessed)
	threadOptimize()
	fmt.Println("Starting thread...")

	prepareInterface()

	go threadUpdateListener()
	go updateStat()

	thread.StandbyRun()
	fmt.Println("Thread started")

	if err := app.Run(); err != nil {
		panic(err)
	}

	thread.Stop()
	fmt.Println("Waiting for result...")
	resultWg.Wait()

	// print result
	if optionOutputType == OUTPUT_TYPE_TXT {
		printToFile()
	}
	if optionOutputType == OUTPUT_TYPE_JSON {
		// TODO
		fmt.Println("JSON output not implemented yet")
	}
	if optionOutputType == OUTPUT_TYPE_CSV {
		// TODO
		fmt.Println("CSV output not implemented yet")
	}

	fmt.Println("Done")
}
