package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/quackduck/aces"
)

var (
	//charset = []rune("zachZACH") //  zachZACH𝐳𝐚𝐜𝐡𝚣𝚊𝚌𝚑𝕫𝕒𝕔𝕙𝘇𝗮𝗰𝗵𝙯𝙖𝙘𝙝𝓏𝒶𝒸𝒽
	charset = []rune("zachZACH𝐳𝐚𝐜𝐡𝚣𝚊𝚌𝚑𝕫𝕒𝕔𝕙𝘇𝗮𝗰𝗵𝙯𝙖𝙘𝙝𝓏𝒶𝒸𝒽𝑧𝑎𝑐ℎ𝒛𝒂𝒄𝒉𝗓𝖺𝖼𝗁𝔃𝓪𝓬𝓱⒵⒜⒞⒣ｚａｃｈⓩⓐⓒⓗ🅉🄰🄲🄷") // 𝑧𝑎𝑐ℎ𝒛𝒂𝒄𝒉𝗓𝖺𝖼𝗁𝔃𝓪𝓬𝓱⒵⒜⒞⒣ｚａｃｈⓩⓐⓒⓗ🅉🄰🄲🄷
	//encodeHaHa = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/")
	//numOfBits  = int(math.Log2(float64(len(encodeHaHa))))
	coding *aces.Coding
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
	var err error
	coding, err = aces.NewCoding(charset)
	if err != nil {
		panic(err)
	}

	if os.Args[1] == "import" {
		err := makeNewZvmExecutable(os.Args[2])
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	err = runZvmExecutable(os.Args[1]+".zvm", os.Args[2:]...)
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
	gzw, _ := gzip.NewWriterLevel(pw, gzip.BestCompression)
	go func() {
		_, err = io.Copy(gzw, src)
		if err != nil {
			panic(err)
		}
		err := src.Close()
		if err != nil {
			panic(err)
		}
		err = gzw.Close()
		if err != nil {
			panic(err)
		}
		err = pw.Close()
		if err != nil {
			panic(err)
		}
	}()
	err = coding.Encode(dst, pr)
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
		panic(err)
	}

	if err = toRun.Chmod(0755); err != nil { // rwx r-x r-x
		return err
	}

	pr, pw := io.Pipe()
	go func() {
		err = coding.Decode(pw, src)
		if err != nil {
			panic(err)
		}
		src.Close()
		pw.Close()
	}()
	gzr, err := gzip.NewReader(pr)
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(toRun, gzr)
	if err != nil {
		panic(err)
	}

	toRun.Sync()
	toRun.Close()
	cmd := exec.Command(toRun.Name(), args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	if err = cmd.Start(); err != nil {
		return err
	}
	err = cmd.Wait()
	return err
}
