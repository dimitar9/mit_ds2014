package viewservice

import "net"
import "net/rpc"
import "log"
import "time"
import "sync"
import "os"

type ViewServer struct {
	mu   sync.Mutex
	l    net.Listener
	dead bool
	me   string

	// Your declarations here.
	lastHB      map[string]time.Time
	lastViewNum map[string]uint
	curView     View
	pendingView View

	flag int
}

//
// server Ping RPC handler.
//
func (vs *ViewServer) Ping(args *PingArgs, reply *PingReply) error {
	DPrintf("received Pring from %v, vienumber is %v\n", args.Me, args.Viewnum)
	DPrintf("current primary is %v current backup is %v cur num is %v\n", vs.curView.Primary, vs.curView.Backup, vs.curView.Viewnum)
	DPrintf("pending primary is %v, pending backup is %v.pen num is %v\n", vs.pendingView.Primary, vs.pendingView.Backup, vs.pendingView.Viewnum)
	if vs.curView.Primary == args.Me && vs.curView.Viewnum < args.Viewnum {
		vs.curView.Viewnum = args.Viewnum
	}
	if vs.curView.Primary == "" && vs.curView.Backup == "" && vs.pendingView.Primary == "" && vs.pendingView.Backup == "" {

		vs.curView.Primary = args.Me
		vs.curView.Viewnum = 1
		DPrintf("Ping in case 0\n")
	} else if vs.lastViewNum[vs.curView.Primary] == vs.curView.Viewnum && vs.flag == 1 && (vs.pendingView.Primary != "" && vs.curView.Primary == args.Me) || (vs.pendingView.Backup != "" && vs.curView.Primary == args.Me) {
		DPrintf("Ping in case 1\n")
        if vs.pendingView.Viewnum == 0 {
			DPrintf("ignoreing switch because backup is not initizliaed.(in Ping)\n")
		} else {
			//handle switch over case
			vs.flag = 0
			vs.curView = vs.pendingView
			vs.curView.Viewnum = args.Viewnum+1
			vs.pendingView.Backup = ""
			vs.pendingView.Primary = ""
			vs.pendingView.Viewnum = 0
			DPrintf("switch happening!\n")
			DPrintf("flag pending is cleared\n")
		}
	} else if vs.curView.Primary != "" && vs.curView.Primary == args.Me && vs.lastViewNum[args.Me] > args.Viewnum {
		DPrintf("Ping in case 2\n")
		DPrintf("vs.lastViewNumber is %v, argsViewnum is %v", vs.lastViewNum[args.Me], args.Viewnum)

		vs.flag = 1
		vs.pendingView.Primary = vs.curView.Backup
		vs.pendingView.Backup = vs.curView.Primary
		vs.pendingView.Viewnum = vs.curView.Viewnum
		DPrintf("Restarted Primary met.\n")

	} else if vs.curView.Primary == "" && vs.pendingView.Primary == "" {
		DPrintf("Ping in case 3\n")
		vs.pendingView.Primary = args.Me
		vs.pendingView.Backup = vs.curView.Backup
		vs.pendingView.Viewnum = vs.curView.Viewnum + 1
		vs.flag = 1
		DPrintf("Initial state, got first ping. Set pending primary and viewnum.\n")

	} else if vs.pendingView.Backup == "" && vs.curView.Backup == "" && vs.curView.Primary != args.Me {

			DPrintf("Ping in case 4\n")
			vs.pendingView.Backup = args.Me
			vs.pendingView.Primary = vs.curView.Primary
			vs.pendingView.Viewnum = vs.curView.Viewnum+1
			DPrintf("initial setup for Backup.\n")

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
	DPrintf("Get() called. vs.curView.viewNum is %v. \n", vs.curView.Viewnum)
	reply.View = vs.curView
	return nil
}

//
// tick() is called once per PingInterval; it should notice
// if servers have died or recovered, and change the view
// accordingly.
//
func (vs *ViewServer) tick() {
	DPrintf("Get in tick, current vienum is %v\n", vs.curView.Viewnum)
	if vs.curView.Backup != "" {
		if time.Now().Sub(vs.lastHB[vs.curView.Backup]) > DeadPings*PingInterval {
			DPrintf("backup timed out. set pending backup to empty\n")
			vs.curView.Backup=""
			vs.pendingView.Primary = vs.curView.Primary
			vs.pendingView.Backup = ""
			vs.pendingView.Viewnum = vs.curView.Viewnum
			vs.flag = 1
		}
	}

	if vs.curView.Primary != "" {
		if time.Now().Sub(vs.lastHB[vs.curView.Primary]) > DeadPings*PingInterval {
			//vs.curView.Primary=""
			if vs.curView.Backup != "" {
				if vs.flag == 0 && vs.lastViewNum[vs.curView.Primary] == vs.curView.Viewnum {
					if vs.lastViewNum[vs.curView.Backup]==0 {
						DPrintf("ignoreing switch because backup is not initizliaed.(in Ping)\n")
					} else {
						vs.curView.Primary = vs.curView.Backup
						vs.curView.Backup = ""
						vs.curView.Viewnum++
						DPrintf("Primary Timedout! switch immediately because flag is 0.\n")
						DPrintf("new cur viewnumber is %v\n", vs.curView.Viewnum)
					}
				} else {

					//change pending view before ack.
					vs.pendingView.Primary = vs.curView.Backup
					vs.pendingView.Backup = ""
					vs.flag = 1
					vs.pendingView.Viewnum = vs.curView.Viewnum + 1
					DPrintf("Primary Timedout! Change pending view.\n make its backup empty because bakcup went to primary current primary is still %v.\n", vs.curView.Primary)
				}
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
	vs.flag = 0

	// tell net/rpc about our RPC server and handlers.
	rpcs := rpc.NewServer()
	rpcs.Register(vs)

	// prepare to receive connections from clients.
	// change "unix" to "tcp" to use over a network.
	os.Remove(vs.me) // only needed for "unix"
	l, e := net.Listen("unix", vs.me)
	if e != nil {
		log.Fatal("listen error: ", e)
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
				DPrintf("ViewServer(%v) accept: %v\n", me, err.Error())
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
