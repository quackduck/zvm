package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"github.com/bearmini/bitstream-go"
	"io"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var (
	//encodeHaHa = []rune("zachZACH𝐳𝐚𝐜𝐡𝚣𝚊𝚌𝚑") //  zachZACH𝐳𝐚𝐜𝐡𝚣𝚊𝚌𝚑𝕫𝕒𝕔𝕙𝘇𝗮𝗰𝗵𝙯𝙖𝙘𝙝𝓏𝒶𝒸𝒽
	encodeHaHa = []rune("zachZACH𝐳𝐚𝐜𝐡𝚣𝚊𝚌𝚑𝕫𝕒𝕔𝕙𝘇𝗮𝗰𝗵𝙯𝙖𝙘𝙝𝓏𝒶𝒸𝒽𝑧𝑎𝑐ℎ𝒛𝒂𝒄𝒉𝗓𝖺𝖼𝗁𝔃𝓪𝓬𝓱⒵⒜⒞⒣ｚａｃｈⓩⓐⓒⓗ🅉🄰🄲🄷") // 𝑧𝑎𝑐ℎ𝒛𝒂𝒄𝒉𝗓𝖺𝖼𝗁𝔃𝓪𝓬𝓱⒵⒜⒞⒣ｚａｃｈⓩⓐⓒⓗ🅉🄰🄲🄷
	numOfBits  = int(math.Log2(float64(len(encodeHaHa))))

	//currProcess *os.Process
)

func main() {
	if os.Args[1] == "--help" {
		fmt.Println(`Hey!!
Import commands like this: zvm import <file>
Run commands like this: zvm <zvm file name without .zvm)

Example session:
   cp $(which ls) ls
   zvm import ls
   zvm ls -l`)
		return
	}
	rand.Seed(time.Now().UnixNano())
	if os.Args[1] == "import" {
		err := makeNewZvmExecutable(os.Args[2])
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	//
	//s := make(chan os.Signal, 1)
	//signal.Notify(s)
	//go func() {
	//	for sig := range s {
	//		if currProcess != nil {
	//			//fmt.Println("signaling")
	//			currProcess.Signal(sig)
	//			//fmt.Println(err)
	//		}
	//	}
	//}()

	err := runZvmExecutable(os.Args[1]+".zvm", os.Args[2:]...)
	if err != nil {
		fmt.Println(err)
	}
}

func makeNewZvmExecutable(exe string) error {
	dst, err := os.Create(filepath.Base(exe) + ".zvm")
	if err != nil {
		return err
	}
	src, err := os.Open(exe)
	if err != nil {
		return err
	}
	pr, pw := io.Pipe()
	go func() {
		// src -> gzip -> pw -> pr -> encode -> dst
		w, _ := gzip.NewWriterLevel(pw, gzip.BestCompression)
		_, err = io.Copy(w, src)
		if err != nil {
			panic(err)
		}
		err := w.Close()
		if err != nil {
			panic(err)
		}
		pw.Close()
	}()
	err = encode(dst, pr)
	if err != nil {
		panic(err)
	}
	return nil
}

func runZvmExecutable(zvme string, args ...string) error {
	toRun, err := os.CreateTemp("", "."+filepath.Base(zvme)+"*.temp.zvm")
	if err != nil {
		return err
	}
	defer os.Remove(toRun.Name())
	src, err := os.Open(zvme)
	if err != nil {
		return err
	}

	if err = toRun.Chmod(0755); err != nil { // rwx r-x r-x
		return err
	}

	pr, pw := io.Pipe()
	go func() {
		err = decode(pw, src)
		if err != nil {
			panic(err)
		}
		src.Close()
		pw.Close()
	}()
	r, err := gzip.NewReader(pr)
	if err != nil {
		return err
	}
	_, err = io.Copy(toRun, r)
	if err != nil {
		return err
	}
	toRun.Sync()
	toRun.Close()
	cmd := exec.Command(toRun.Name(), args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	if err = cmd.Start(); err != nil {
		return err
	}

	//currProcess = cmd.Process
	//s := make(chan os.Signal, 4)
	//signal.Notify(s)
	//go func() {
	//	for _ = range s {
	//		fmt.Println("I WILL NOT STOP")
	//		time.AfterFunc(time.Second*5, func() {
	//			os.Exit(0)
	//		})
	//		//cmd.Process.Signal(sig)
	//	}
	//}()
	err = cmd.Wait()
	return err
}

func encode(dst io.Writer, src io.Reader) error {

	r := bitstream.NewReader(src, nil)
	res := make([]byte, 0, 1024)
	for {
		chunk, err := r.ReadNBitsAsUint8(uint8(numOfBits))
		if err != nil {
			if err == io.EOF {
				_, err = dst.Write(res)
			}
			return err
		}
		res = append(res, string(encodeHaHa[chunk])...)
		if len(res) > 1024 {
			n, err := dst.Write(res)
			if err != nil {
				return err
			}
			if n != len(res) {
				panic("whoa")
			}
			res = make([]byte, 0, 1024)
		}
	}
}

func decode(dst io.Writer, src io.Reader) error {
	w := bitstream.NewWriter(dst)
	br := bufio.NewReader(src)
	for {
		r, _, err := br.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		for i, char := range encodeHaHa {
			if r == char {
				err := w.WriteNBitsOfUint8(uint8(numOfBits), byte(i))
				if err != nil {
					return err
				}
			}
		}
	}
	return w.Flush()
}
