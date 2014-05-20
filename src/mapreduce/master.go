package mapreduce
import "container/list"
import "fmt"



type WorkerInfo struct {
  address string
  // You can add definitions here.
}


// Clean up all workers by sending a Shutdown RPC to each one of them Collect
// the number of jobs each work has performed.
func (mr *MapReduce) KillWorkers() *list.List {
  l := list.New()
  for _, w := range mr.Workers {
    DPrintf("DoWork: shutdown %s\n", w.address)
    args := &ShutdownArgs{}
    var reply ShutdownReply;
    ok := call(w.address, "Worker.Shutdown", args, &reply)
    if ok == false {
      fmt.Printf("DoWork: RPC %s shutdown error\n", w.address)
    } else {
      l.PushBack(reply.Njobs)
    }
  }
  return l
}

func (mr *MapReduce) RunMaster() *list.List {
	DPrintf("Runmaster.\n")
  // Your code here


  args := &DoJobArgs{"824-mrinput.txt","DoMap",0,5}
  var reply DoJobReply
  var err bool

  err = call(mr.Workers["worker0"].address, "Worker.DoJob", args, &reply)
  if err  {
  	fmt.Println("wk error:%d",err)
  }
  return mr.KillWorkers()
}
