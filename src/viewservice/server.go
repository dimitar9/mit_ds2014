
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
  fmt.Printf("received Pring from %v, vienumber is %v\n", args.Me, args.Viewnum)
  fmt.Printf("current primary is %v current backup is %v\n", vs.curView.Primary, vs.curView.Backup)
  fmt.Printf("pending primary is %v, pending backup is %v.\n",vs.pendingView.Primary,vs.pendingView.Backup)
  if (vs.pendingView.Primary != "" && vs.pendingView.Primary==args.Me)  ||(vs.pendingView.Backup!="" && vs.curView.Primary==args.Me) {
	  fmt.Printf("Ping in case 1\n")
	  //handle switch over case
	  vs.flag = 0
	  vs.curView = vs.pendingView
	  vs.pendingView.Backup=""
	  vs.pendingView.Primary=""
	  vs.pendingView.Viewnum=0
	  fmt.Printf("switch happening!\n")
	  fmt.Printf("flag pending is cleared\n")
  } else if vs.curView.Primary != "" && vs.curView.Primary == args.Me &&  vs.lastViewNum[args.Me] > args.Viewnum {
	  fmt.Printf("Ping in case 2\n")
	  vs.flag = 1
	  vs.pendingView.Primary=vs.curView.Primary //old primary
	  vs.pendingView.Backup=vs.curView.Backup//old backup
	  vs.curView.Primary = ""
	  vs.curView.Primary = vs.curView.Backup
	  vs.curView.Backup = ""
	  fmt.Printf("REALONE!!!! flag pending is 1, waiting for ack from %v to clear flag.\n", vs.curView.Primary)

  } else if vs.curView.Primary == "" && vs.pendingView.Primary== ""{
	  fmt.Printf("Ping in case 3\n")
	  vs.pendingView.Primary = args.Me
	  vs.pendingView.Backup = vs.curView.Backup
	  vs.pendingView.Viewnum = vs.curView.Viewnum+1
	  vs.flag = 1
	  fmt.Println("Initial state, got first ping. Set pending primary and viewnum.")

  } else if vs.pendingView.Backup == ""  && vs.curView.Backup=="" && vs.curView.Primary != args.Me {
	  fmt.Printf("Ping in case 4\n")
	  vs.pendingView.Backup = args.Me
	  vs.pendingView.Primary=vs.curView.Primary
	  vs.pendingView.Viewnum = vs.curView.Viewnum+1
	  fmt.Println("initial setup for Backup.")
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
	fmt.Println("Get in tick, current vienum is ", vs.curView.Viewnum)
	if vs.curView.Backup != "" {
		if time.Now().Sub(vs.lastHB[vs.curView.Backup]) > DeadPings*PingInterval {
			fmt.Println("backup timed out. set pending backup to empty")
			vs.pendingView.Backup = ""
			vs.flag = 1
		}
	}

	if vs.curView.Primary != "" {
		if time.Now().Sub(vs.lastHB[vs.curView.Primary]) > DeadPings*PingInterval {
			vs.curView.Primary=""
			if vs.curView.Backup != "" {

				//change pending view before ack.
				vs.pendingView.Primary = vs.curView.Backup
				vs.pendingView.Backup=""
				vs.flag = 1
				vs.pendingView.Viewnum =vs.curView.Viewnum+1
				fmt.Printf("Primary Timedout! Change pending view.\n make its backup empty because bakcup went to primary current primary is still %v.\n", vs.curView.Primary)
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
