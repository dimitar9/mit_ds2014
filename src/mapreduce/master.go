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

  //r := new (WorkerInfo)  
  var worker_str string
  worker_str = <- mr.registerChannel 
  fmt.Println("Initialize map reduce!!!!!!!")
  fmt.Printf("worker_str is %s \n",worker_str)
  //(*r).address =port(worker_str)
 // mr.Workers[worker_str]= r

  args := &DoJobArgs{"824-mrinput.txt","Map",0,5}
  var reply DoJobReply
  var err bool

  err = call(worker_str, "Worker.DoJob", args, &reply)
  if err  {
  	fmt.Println("wk error:%d",err)
  }
  return mr.KillWorkers()
}
