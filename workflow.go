package flowmat

import (
  // "log"
)

type Processer interface {
  Name() string
  Run()
  SetIn(string, chan interface{})
  SetOut(string, chan interface{})
}

type Workflow struct {
  name string
  proc map[string]Processer
  inPorts map[string]*Port
  outPorts map[string]*Port
  exposePorts map[string]*Port
}

// NewWorkflow create workflow object
func NewWorkflow(name string) *Workflow {
  wf := &Workflow{
    name: name,
    proc: make(map[string]Processer),
    inPorts: make(map[string]*Port),
    outPorts: make(map[string]*Port),
    exposePorts: make(map[string]*Port),
  }
  return wf
}

func (w *Workflow) Name() string {
  return w.name
}

// Add process to workflow list
func (w *Workflow) Add(p Processer) {
  w.proc[p.Name()] = p
}

// Connect outport of Process A(sendProc) to inport of Process B(recvProc)
func (w *Workflow) Connect(sendProc, sendPort, recvProc, recvPort string) {
  s := w.proc[sendProc]
  r := w.proc[recvProc]
  out := make(chan interface{})
  in := make(chan interface{})

  s.SetOut(sendPort, out)
  r.SetIn(recvPort, in)

  go func() {
    v := <-out
    in <- v
  }()
}

func (w *Workflow) SetIn(name string, channel chan interface{}) {
  w.inPorts[name] = &Port{
    channel: channel,
  }
}

// SetOut bind port to a channel
func (w *Workflow) SetOut(name string, channel chan interface{}) {
  w.outPorts[name] = &Port{
    channel: channel,
  }
}

// MapIn map inPorts of process to workflow
func (w *Workflow) MapIn(name, procName, portName string) {
  channel := make(chan interface{})
  w.SetIn(name, channel)

  p := w.proc[procName]
  p.SetIn(portName, channel)
}

// MapOut map outPorts of process to workflow
func (w *Workflow) MapOut(name, procName, portName string) {
  channel := make(chan interface{})
  w.SetOut(name, channel)

  p := w.proc[procName]
  p.SetOut(portName, channel)
}

func (w *Workflow) ExposeIn(name string) chan interface{} {
  w.exposePorts[name] = new(Port)
  return w.inPorts[name].channel
}

func (w *Workflow) ExposeOut(name string) chan interface{} {
  w.exposePorts[name] = new(Port)
  return w.outPorts[name].channel
}

// In pass the data to the inport
func (w *Workflow) In(portName string, data interface{}) {
  port := w.inPorts[portName]
  port.cache = data
}

// Out get the result from outport
func (w *Workflow) Out(portName string) interface{} {
  data := w.outPorts[portName].cache
  // if data == nil {
  //   log.Panicf("%s has not get data", portName)
  // }
  return data
}

// Run the workflow aka its process in order
func (w *Workflow) Run() {
  for _, p := range w.proc {
    p.Run()
  }
  for name, port := range w.inPorts {
    cacheData := port.cache
    if _, ok := w.exposePorts[name]; !ok {
      port.channel <- cacheData
    }
  }
  // if the port not expose, store it in cache
  for name, port := range w.outPorts {
    if _, ok := w.exposePorts[name]; !ok {
      data := <-port.channel
      port.cache = data
    }
  }
}
