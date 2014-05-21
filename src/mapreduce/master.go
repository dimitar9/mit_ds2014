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

  var worker_str string
  
  //loop thru Map workers

    
  worker_str = <- mr.registerChannel
  DPrintf("Worker_str is %s \n",worker_str)

  args := &DoJobArgs{mr.file,"Map",0,1}//mr.nReduce}
  var reply DoJobReply
  var ret bool
  ret = call(worker_str, "Worker.DoJob", args, &reply)
  if ret  {
  	fmt.Println("wk worker done.\n")
  } else
  {
    fmt.Println("wk worker fail.\n")
  }
  DPrintf("map finished.")


  args_reduce := &DoJobArgs{mr.file,"Reduce",0,1}//mr.nMap}
  ret = call(worker_str, "Worker.DoJob", args_reduce, &reply)
  if ret  {
    fmt.Println("wk reduce worker done.\n")
  } else
  {
    fmt.Println("wk reduce worker fail.\n")
  }

  DPrintf("reduce finished.")
  return mr.KillWorkers()
}
