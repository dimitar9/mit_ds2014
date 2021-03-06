package mapreduce
import "container/list"
import "fmt"

//http://css.csail.mit.edu/6.824/2014/labs/lab-1.html

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

  // var worker_str string
  
  //loop thru Map workers
  for i := 0; i < mr.nMap; i++ {
    DPrintf("worker number is %d\n", mr.workerNumber)
    var sucess bool
    sucess = false
    for ; sucess != true;  {
      worker_str := <- mr.registerChannel

      DPrintf("Worker_str is %s \n",worker_str)

      args := &DoJobArgs{mr.file,"Map",i,mr.nReduce}
      var reply DoJobReply
      var ret bool
      ret = call(worker_str, "Worker.DoJob", args, &reply)
      if ret  {
      	fmt.Println("wk worker done.\n")
        fmt.Println(worker_str)
        mr.registerChannel <- worker_str
        sucess = true
      } else
      {
        fmt.Println("wk workerfail.\n")
      }
    }
    DPrintf("map finished.")

  }



  for i := 0; i < mr.nReduce; i++ {
    var reduce_sucess bool
    reduce_sucess = false
    for ; reduce_sucess != true;  {
      worker_str := <- mr.registerChannel

      args_reduce := &DoJobArgs{mr.file,"Reduce",i,mr.nMap}
      var reply DoJobReply
      var ret bool
      ret = call(worker_str, "Worker.DoJob", args_reduce, &reply)
      if ret  {
        fmt.Println("wk reduce worker done.\n")
        mr.registerChannel <- worker_str
        reduce_sucess = true
      } else
      {
        fmt.Println("wk reduce workerfail.\n")
      }
    }
    DPrintf("reduce finished.")

    
  }

  return mr.KillWorkers()
}
