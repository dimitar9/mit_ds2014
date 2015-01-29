
package viewservice

import "net"
import "net/rpc"
import "log"
import "time"
import "sync"
import "fmt"
import "os"

type ViewServer struct {
  mu sync.Mutex
  l net.Listener
  dead bool
  me string


  // Your declarations here.
  lastHB map[string]time.Time
  lastViewNum map[string]uint
  curView View
  pendingView View

  flag int
}

//
// server Ping RPC handler.
//
func (vs *ViewServer) Ping(args *PingArgs, reply *PingReply) error {

  if vs.curView.Primary != "" {
	  if vs.curView.Primary == args.Me {

			  vs.pendingView.Viewnum= args.Viewnum + 1//vs.curView.Viewnum+1
			  fmt.Printf("pending view viewnum is (%v), curView viewnumber is (%v).\n", vs.pendingView.Viewnum, vs.curView.Viewnum)
			  vs.curView.Viewnum = vs.pendingView.Viewnum
			  fmt.Printf("view switch!\n")

		  vs.flag = 0
	  }
  }
  // Your code here.
  if vs.curView.Primary == "" && vs.pendingView.Primary== ""{
	  vs.curView.Primary = args.Me

	  vs.curView.Viewnum = vs.curView.Viewnum+1
	  fmt.Println("@@@@increase view Num in ping.Primary")
  } else if vs.curView.Backup == ""  && vs.curView.Primary != args.Me {
	  vs.curView.Backup = args.Me
	  vs.pendingView.Viewnum = vs.curView.Viewnum+1
	  fmt.Println("@@@@increase view Num in ping Backup")
  }
  vs.lastHB[args.Me] = time.Now()
  vs.lastViewNum[args.Me] = args.Viewnum
  reply.View = vs.curView
  return nil
}

// 
// server Get() RPC handler.
//
func (vs *ViewServer) Get(args *GetArgs, reply *GetReply) error {

  // Your code here.
  reply.View = vs.curView
  return nil
}


//
// tick() is called once per PingInterval; it should notice
// if servers have died or recovered, and change the view
// accordingly.
//
func (vs *ViewServer) tick() {
	fmt.Println("@@@@in tick current vienum is ", vs.curView.Viewnum)
	// Your code here.
//	fmt.Printf("lastHb curViewPrimary (%v) time now: %v\n", vs.lastHB[vs.curView.Primary], time.Now())
//	fmt.Printf("lastHb curViewBackup (%v) time now: %v\n", vs.lastHB[vs.curView.Backup], time.Now())
//    fmt.Printf("diff time is (%v)",time.Now().Sub(vs.lastHB[vs.curView.Primary]))

	if vs.curView.Backup != "" {
		if time.Now().Sub(vs.lastHB[vs.curView.Backup]) > DeadPings*PingInterval {
			vs.curView.Backup = ""
			if vs.flag == 0 {
				vs.pendingView.Viewnum = vs.curView.Viewnum+1
				fmt.Println("@@@@increase view Num in curview.Backup")
				vs.flag = 1
			}

		}
	}

	if vs.curView.Primary != "" {
		if time.Now().Sub(vs.lastHB[vs.curView.Primary]) > DeadPings*PingInterval {
			if vs.curView.Backup != "" {
				vs.curView.Primary = vs.curView.Backup

				vs.curView.Backup = ""
				fmt.Println("make backup empty because bakcup went to primary.\n")
			}
			if vs.flag == 0 {
				vs.pendingView.Viewnum= vs.curView.Viewnum+1
				fmt.Println("@@@@increase view Num in curview.Primary")
				vs.flag = 1
			}
		}
	}

}

//
// tell the server to shut itself down.
// for testing.
// please don't change this function.
//
func (vs *ViewServer) Kill() {
  vs.dead = true
  vs.l.Close()
}

func StartServer(me string) *ViewServer {
  vs := new(ViewServer)
  vs.me = me
  // Your vs.* initializations here.
  vs.lastHB = make(map[string]time.Time)
  vs.lastViewNum = make(map[string]uint)
  vs.curView = View{Viewnum: 0, Primary: "", Backup: ""}
  vs.pendingView = View{Viewnum: 0, Primary: "", Backup: ""}
  vs.flag=0

  // tell net/rpc about our RPC server and handlers.
  rpcs := rpc.NewServer()
  rpcs.Register(vs)

  // prepare to receive connections from clients.
  // change "unix" to "tcp" to use over a network.
  os.Remove(vs.me) // only needed for "unix"
  l, e := net.Listen("unix", vs.me);
  if e != nil {
    log.Fatal("listen error: ", e);
  }
  vs.l = l

  // please don't change any of the following code,
  // or do anything to subvert it.

  // create a thread to accept RPC connections from clients.
  go func() {
    for vs.dead == false {
      conn, err := vs.l.Accept()
      if err == nil && vs.dead == false {
        go rpcs.ServeConn(conn)
      } else if err == nil {
        conn.Close()
      }
      if err != nil && vs.dead == false {
        fmt.Printf("ViewServer(%v) accept: %v\n", me, err.Error())
        vs.Kill()
      }
    }
  }()

  // create a thread to call tick() periodically.
  go func() {
    for vs.dead == false {
      vs.tick()
      time.Sleep(PingInterval)
    }
  }()

  return vs
}
