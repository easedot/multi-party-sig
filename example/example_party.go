package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/taurusgroup/multi-party-sig/internal/test"
	"github.com/taurusgroup/multi-party-sig/pkg/ecdsa"
	"github.com/taurusgroup/multi-party-sig/pkg/math/curve"
	"github.com/taurusgroup/multi-party-sig/pkg/party"
	"github.com/taurusgroup/multi-party-sig/pkg/pool"
	"github.com/taurusgroup/multi-party-sig/pkg/protocol"
	"github.com/taurusgroup/multi-party-sig/protocols/cmp"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func CMPKeygen1(id party.ID, ids party.IDSlice, threshold int, n test.INetwork, pl *pool.Pool) (*cmp.Config, error) {
	h, err := protocol.NewMultiHandler(cmp.Keygen(curve.Secp256k1{}, id, ids, threshold, pl), nil)
	if err != nil {
		return nil, err
	}
	test.HandlerLoop(id, h, n)
	r, err := h.Result()
	if err != nil {
		return nil, err
	}

	return r.(*cmp.Config), nil
}

func CMPRefresh1(c *cmp.Config, n test.INetwork, pl *pool.Pool) (*cmp.Config, error) {
	hRefresh, err := protocol.NewMultiHandler(cmp.Refresh(c, pl), nil)
	if err != nil {
		return nil, err
	}
	test.HandlerLoop(c.ID, hRefresh, n)

	r, err := hRefresh.Result()
	if err != nil {
		return nil, err
	}

	return r.(*cmp.Config), nil
}

func CMPSign1(c *cmp.Config, m []byte, signers party.IDSlice, n test.INetwork, pl *pool.Pool) error {
	h, err := protocol.NewMultiHandler(cmp.Sign(c, signers, m, pl), nil)
	if err != nil {
		return err
	}
	test.HandlerLoop(c.ID, h, n)

	signResult, err := h.Result()
	if err != nil {
		return err
	}

	signature := signResult.(*ecdsa.Signature)

	//c.PublicPoint 公钥 m 文本 signature签名
	log.Printf("msg:%s", m)
	prvSig, _ := signature.R.MarshalBinary()
	log.Printf("prvSig:%s ", prvSig)
	if !signature.Verify(c.PublicPoint(), m) {
		return errors.New("failed to verify cmp signature")
	}
	return nil
}

func mustCopy(dst net.Conn, src io.Reader) {
	if _, err := io.Copy(dst, src); err != nil {
		log.Println("Conn Error:", err)
	}
}

func processCmd(uid party.ID, tcp *test.NetworkBroadCast, ids party.IDSlice, threshold int, group curve.Secp256k1) {
	f := fmt.Sprintf("config_%s.txt", uid)
	pl := pool.NewPool(0)
	defer pl.TearDown()
	for cmd := range tcp.CMD() {
		log.Println("CMD", cmd)
		switch cmd {
		case "all":
			//var wg sync.WaitGroup
			//messageToSign := []byte("hello")
			//All(uid, ids, threshold, messageToSign, tcp, &wg, pl)
		case "sign":
			signers := ids[:threshold+1]
			messageToSign := []byte("hello")
			cfg, err := loadConfig(f, group)
			err = CMPSign1(cfg, messageToSign, signers, tcp, pl)
			if err != nil {
				log.Print("Error", err)
			}
			log.Print("Sign OK")

		case "addr":
			cfgOld, err := loadConfig(f, group)
			if err != nil {
				log.Println("load config error:", err)
			}

			eth1 := "m/44/60/0/0/1"
			genAddress(cfgOld, eth1)
			eth2 := "m/44/60/0/0/2"
			genAddress(cfgOld, eth2)

		case "refresh":
			cfgOld, err := loadConfig(f, group)
			if err != nil {
				log.Println("load config error:", err)
			}

			cfgNew, err := CMPRefresh1(cfgOld, tcp, pl)
			if err != nil {
				log.Println("load refresh error:", err)
			}

			cfjo, _ := json.Marshal(cfgOld)
			cfjn, _ := json.Marshal(cfgNew)
			log.Printf("refresh old cfg:%s", cfjo)
			log.Printf("refresh new cfg:%s", cfjn)

			pa1, err := cfgOld.PublicPoint().MarshalBinary()
			log.Printf("Old public bip32 compress:%x", pa1)
			pa2, err := cfgNew.PublicPoint().MarshalBinary()
			log.Printf("New public bip32 compress:%x", pa2)

		case "gen":
			cfg, err := CMPKeygen1(uid, ids, threshold, tcp, pl)
			if err != nil {
				log.Println("Error:", err)
			}
			cfj, _ := json.Marshal(cfg)
			log.Printf("cfg:%s", cfj)

			writeConfig(cfg, f)

		}
	}
}

func genAddress(cfgOld *cmp.Config, path string) {
	cfgNew, err := cfgOld.DeriveBIP44(path)
	if err != nil {
		log.Println("DeriveBipError", err)
	}
	log.Printf("BIP44:Path: %s", path)
	pubkey, err := cfgNew.PublicPoint().MarshalBinary()
	log.Printf("pubkey compress:%x", pubkey)
}

func writeConfig(cfg *cmp.Config, f string) {
	cfb, _ := cfg.MarshalBinary()
	ioutil.WriteFile(f, cfb, 0644)
}

func loadConfig(f string, group curve.Secp256k1) (*cmp.Config, error) {
	cfb, err := ioutil.ReadFile(f)
	if err != nil {
		log.Println("Error", err)
	}
	//debug
	//log.Printf("Config Content:%s", cfb)
	cfg := cmp.EmptyConfig(group)
	err = cfg.UnmarshalBinary(cfb)
	if err != nil {
		log.Println("Error", err)
	}
	return cfg, err
}

func main() {
	id := flag.Int("id", 0, "party id")
	srv := flag.String("srv", "localhost:8000", "connect to tcp server")
	flag.Parse()

	ids := party.IDSlice{"a", "b", "c"}
	threshold := 2
	group := curve.Secp256k1{}

	conn, err := net.Dial("tcp", *srv)
	if err != nil {
		log.Fatal(err)
	}

	uid := ids[*id]
	log.Println("Start party:", uid)

	tcp := test.NewNetworkTcp(uid, conn)

	go processCmd(uid, tcp, ids, threshold, group)

	go mustCopy(conn, os.Stdin)

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	conn.Close()
	log.Print("Shutdown Server ...")

}
