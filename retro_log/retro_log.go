package retro_log

import (
	"fmt"
	"github.com/pigpaxos/pigpaxos/hlc"
	"github.com/pigpaxos/pigpaxos/log"
	"os"
	"strconv"
	"strings"
	"sync"
)

type RetroVar struct {
	name	*string
	valstr	string
}

func (v *RetroVar) varToStr() string {
	if v.name == nil {
		return v.valstr
	}
	vstr := *v.name + ":" + v.valstr
	return vstr
}

type RqlStruct struct {
	name *string
	vars []*RetroVar

	sync.RWMutex
}

func NewRqlStruct(structName *string) *RqlStruct {
	return &RqlStruct{name: structName, vars: make([]*RetroVar ,0)}
}

func (s *RqlStruct) structToStr() string {
	s.RLock()
	defer s.RUnlock()
	var sb strings.Builder
	sb.Reset()
	if s.name != nil {
		sb.WriteString(*s.name)
		sb.WriteString(":")
	}

	sb.WriteString("[")
	for i, v := range s.vars {
		sb.WriteString(v.varToStr())
		if i + 1 < len(s.vars) {
			sb.WriteString(",")
		}
	}
	sb.WriteString("]")

	return sb.String()
}

func (s *RqlStruct) addVar(v *RetroVar) {
	s.Lock()
	defer s.Unlock()
	s.vars = append(s.vars, v)
}

func (s *RqlStruct) AddVarStr(name, val string) *RqlStruct {
	s.addVar(&RetroVar{name: &name, valstr: "\"" + val + "\""})
	return s
}

func (s *RqlStruct) AddVarInt32(name string, val int) *RqlStruct {
	strval := fmt.Sprintf("%d", val)
	s.addVar(&RetroVar{name: &name, valstr: strval})
	return s
}

func (s *RqlStruct) AddVarInt(name string, val int64) *RqlStruct {
	strval := fmt.Sprintf("%d", val)
	s.addVar(&RetroVar{name: &name, valstr: strval})
	return s
}

func (s *RqlStruct) AddVarFloat(name string, val float64) *RqlStruct {
	strval := fmt.Sprintf("%f", val)
	s.addVar(&RetroVar{name: &name, valstr: strval})
	return s
}

type RqlSet struct {
	name	*string
	//vars	[]string
	vars	map[string] bool
	isAdd	bool
}

func NewRqlSet(setName string, isAdd bool) *RqlSet {
	return &RqlSet{name: &setName, vars: make(map[string] bool), isAdd: isAdd}
}

func (s *RqlSet) writeRqlSet(sb *strings.Builder, erase bool) {
	sb.WriteString(*s.name)
	sb.WriteString(":")
	if s.isAdd {
		sb.WriteString("...{")
	} else {
		sb.WriteString("--{")
	}

	varCount := 0
	for k, _ := range s.vars {
		sb.WriteString(k)
		varCount++
		if varCount < len(s.vars) {
			sb.WriteString(",")
		}
	}

	sb.WriteString("}")

	if erase {
		s.vars = make(map[string] bool, 0)
	}
}

func (s *RqlSet) addVal(val string) {
	s.vars[val] = true
}

func (s *RqlSet) removeVal(val string) {
	delete(s.vars, val)
}

func (s *RqlSet) addAllVal(val []string) {
	for _, v := range val {
		s.vars[v] = true
	}
}

func (s *RqlSet) removeAllVal(val []string) {
	for _, v := range val {
		delete(s.vars, v)
	}
}

type RetroLog struct {
	name 		string
	nodeIdStr	string
	logfile		*os.File
	sb 			strings.Builder

	ts 			*hlc.Timestamp

	tempVars      []*RetroVar
	tempAddSet    map[string]*RqlSet
	tempRemoveSet map[string]*RqlSet

	snapVars	  map[string]*RetroVar
	snapSets	  map[string]*RqlSet
	expireSets    map[string]*TimerRqlSet

	blockSize	  int
	lastBlockTime int64

	takeSnapshots bool

	sync.RWMutex

	txLock sync.RWMutex
}

func NewRetroLog(name string, id int, logdir string, blockSize int, takeSnapshots bool) *RetroLog {
	pid := os.Getpid()
	f, err := os.OpenFile(logdir + "retro_" + strconv.Itoa(pid) + ".log",
		os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0755)
	if err != nil{
		log.Fatalf("Could not create retroscope log file for pid %v", pid)
	}

	addSet := make(map[string]*RqlSet, 0)
	removeSet := make(map[string]*RqlSet, 0)
	expireSets := make(map[string]*TimerRqlSet, 0)

	snapVars := make(map[string]*RetroVar, 0)
	snapSets := make(map[string]*RqlSet, 0)

	vName := "retrolog_go"
	snapVars[vName] = &RetroVar{name: &vName, valstr: "1"}

	return &RetroLog{
		name:          	name,
		nodeIdStr:    	strconv.Itoa(id),
		logfile:      	f,
		blockSize:		blockSize,
		lastBlockTime:  0,
		tempVars:    	make([]*RetroVar, 0),
		tempAddSet:  	addSet,
		tempRemoveSet: 	removeSet,
		expireSets:    	expireSets,
		snapVars: 	   	snapVars,
	    snapSets:	   	snapSets,
		takeSnapshots:	takeSnapshots}
}

func (l *RetroLog) StartTx() *RetroLog {
	l.txLock.Lock()
	ts := hlc.HLClock.Now()
	l.sb.Reset()
	l.ts = &ts
	if l.takeSnapshots {
		blockTime := l.getBlockTime(l.ts.PhysicalTime)
		if l.lastBlockTime < blockTime {
			// write snapshot
			l.lastBlockTime = blockTime
			l.WriteSnap(blockTime)
		}
	}
	return l
}

func (l *RetroLog) AppendVarStr(varName string, varVal string) *RetroLog {
	l.Lock()
	defer l.Unlock()
	v := &RetroVar{name: &varName, valstr: "\"" + varVal + "\""}
	l.tempVars = append(l.tempVars, v)
	return l
}

func (l *RetroLog) AppendVarInt(varName string, varVal int64) *RetroLog {
	l.Lock()
	defer l.Unlock()
	strval := fmt.Sprintf("%d", varVal)
	v := &RetroVar{name: &varName, valstr: strval}
	l.tempVars = append(l.tempVars, v)
	return l
}

func (l *RetroLog) AppendVarInt32(varName string, varVal int) *RetroLog {
	l.Lock()
	defer l.Unlock()
	strval := fmt.Sprintf("%d", varVal)
	v := &RetroVar{name: &varName, valstr: strval}
	l.tempVars = append(l.tempVars, v)
	return l
}

func (l *RetroLog) AppendVarFloat(varName string, varVal float64) *RetroLog {
	l.Lock()
	defer l.Unlock()
	s := fmt.Sprintf("%f", varVal)
	v := &RetroVar{name: &varName, valstr: s}
	l.tempVars = append(l.tempVars, v)
	//l.snapVars[varName] = v
	return l
}

func (l *RetroLog) CreateTimerSet(setName string, expireTime int64) {
	l.expireSets[setName] = NewTimerRqlSet(setName, expireTime)
}

func (l *RetroLog) appendSet(setName string, setVal string) *RetroLog{
	l.Lock()
	defer l.Unlock()

	expSet, expireOk := l.expireSets[setName]
	if expireOk {
		expSet.Add(setVal, l.ts.PhysicalTime)
	}

	addSet, ok := l.tempAddSet[setName]
	if !ok {
		addSet = NewRqlSet(setName, true)
		l.tempAddSet[setName] = addSet
	}
	addSet.addVal(setVal)

	return l
}

func (l *RetroLog) AppendSetStr(setName string, setVal string) *RetroLog{
	return l.appendSet(setName, "\"" + setVal + "\"")
}

func (l *RetroLog) AppendSetInt32(setName string, setVal int) *RetroLog{
	strval := fmt.Sprintf("%d", setVal)
	return l.appendSet(setName, strval)
}

func (l *RetroLog) AppendSetInt(setName string, setVal int64) *RetroLog{
	strval := fmt.Sprintf("%d", setVal)
	return l.appendSet(setName, strval)
}

func (l *RetroLog) AppendSetFloat(setName string, setVal float64) *RetroLog{
	valStr := fmt.Sprintf("%f", setVal)
	return l.appendSet(setName, valStr)
}

func (l *RetroLog) AppendSetStruct(setName string, setVal *RqlStruct) *RetroLog{
	return l.appendSet(setName, setVal.structToStr())
}

func (l *RetroLog) removeSet(setName string, setVal string) *RetroLog{
	l.Lock()
	defer l.Unlock()
	snapSet, snapOk := l.snapSets[setName]
	if !snapOk {
		snapSet = NewRqlSet(setName, true)
		l.snapSets[setName] = snapSet
	}
	snapSet.removeVal(setVal)

	removeSet, ok := l.tempRemoveSet[setName]
	if !ok {
		removeSet = NewRqlSet(setName, false)
		l.tempAddSet[setName] = removeSet
	}
	removeSet.addVal(setVal)
	return l
}

func (l *RetroLog) RemoveSetStr(setName string, setVal string) *RetroLog{
	return l.removeSet(setName, "\"" + setVal + "\"")
}

func (l *RetroLog) RemoveSetInt(setName string, setVal int64) *RetroLog{
	strval := fmt.Sprintf("%d", setVal)
	return l.removeSet(setName, strval)
}

func (l *RetroLog) RemoveSetFloat(setName string, setVal float64) *RetroLog{
	valStr := fmt.Sprintf("%f", setVal)
	return l.removeSet(setName, valStr)
}

func (l *RetroLog) RemoveSetStruct(setName string, setVal *RqlStruct) *RetroLog{
	return l.removeSet(setName, setVal.structToStr())
}

func (l *RetroLog) WriteSnap(blockTime int64) {
	ts := *hlc.NewTimestamp(blockTime, 0)
	l.writeHeader(&ts, true)
	ok := l.writeSnapVars()
	if ok {
		l.sb.WriteString(",")
	}

	l.writeSnapSets()

	line := strings.Trim(l.sb.String(), ",")
	l.logfile.WriteString(line + "\n")
	l.sb.Reset()
}

func (l *RetroLog) Commit() {
	l.Lock()
	defer l.Unlock()
	defer l.txLock.Unlock()

	if l.ts == nil{
		log.Error("Cannot commit RetroLog. Transaction was not started")
	}

	l.writeHeader(l.ts, false)
	ok := l.writeVars()
	if ok {
		l.sb.WriteString(",")
	}

	l.writeSets()

	line := strings.Trim(l.sb.String(), ",")
	l.logfile.WriteString(line + "\n")
	l.ts = nil
}

func (l *RetroLog) getBlockTime(time int64) int64 {
	return time / int64(l.blockSize) * int64(l.blockSize)
}

func (l *RetroLog) writeHeader(ts *hlc.Timestamp, isSnapshot bool) {
	if isSnapshot {
		l.sb.WriteString("<snapt")
	} else {
		l.sb.WriteString("<t")
	}
	l.sb.WriteString(strconv.FormatInt(ts.PhysicalTime, 10))
	l.sb.WriteString(",")
	l.sb.WriteString(strconv.FormatInt(ts.ToInt64(), 10))
	l.sb.WriteString(",")
	l.sb.WriteString(l.nodeIdStr)
	l.sb.WriteString(">:")
}

func (l *RetroLog) writeSnapVars() bool {
	writtenVals := len(l.snapVars) > 0
	writtenValsCount := 1
	for _, v := range l.snapVars {
		l.sb.WriteString(v.varToStr())
		if writtenValsCount < len(l.tempVars) - 1 {
			l.sb.WriteString(",")
		}
		writtenValsCount++
	}
	return writtenVals
}

func (l *RetroLog) writeVars() bool {
	writtenVals := len(l.tempVars) > 0
	for i, v := range l.tempVars {
		l.sb.WriteString(v.varToStr())
		if i < len(l.tempVars) - 1 {
			l.sb.WriteString(",")
		}
	}

	l.tempVars = l.tempVars[:0]
	return writtenVals
}

func (l *RetroLog) writeSnapSets() {
	for _, s := range l.snapSets {
		if len(s.vars) > 0 {
			s.writeRqlSet(&l.sb, true)
			l.sb.WriteString(",")
		}
	}

	for _, es := range l.expireSets {
		s := es.Snapshot()
		if len(s.vars) > 0 {
			s.writeRqlSet(&l.sb, false)
			l.sb.WriteString(",")
		}
	}
}

func (l *RetroLog) writeSets() {
	for _, s := range l.tempAddSet {
		if len(s.vars) > 0 {
			s.writeRqlSet(&l.sb, true)
			l.sb.WriteString(",")
		}
	}

	for _, s := range l.tempRemoveSet {
		if len(s.vars) > 0 {
			s.writeRqlSet(&l.sb, true)
			l.sb.WriteString(",")
		}
	}

	for _, es := range l.expireSets {
		s := es.ExpireItems(l.ts.PhysicalTime)
		if len(s.vars) > 0 {
			s.writeRqlSet(&l.sb, true)
			l.sb.WriteString(",")
		}
	}
}