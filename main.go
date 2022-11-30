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
		err = makeNewZvmExecutable(os.Args[2])
		if err != nil {
			fmt.Println("error:", err)
		}
		return
	}

	err = runZvmExecutable(os.Args[1]+".zvm", os.Args[2:]...)
	if err != nil {
		fmt.Println("error:", err)
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

	errch := make(chan error)
	go func() {
		_, err := io.Copy(gzw, src)
		if err != nil {
			errch <- err
			return
		}
		err = src.Close()
		if err != nil {
			errch <- err
			return
		}
		err = gzw.Close()
		if err != nil {
			errch <- err
			return
		}
		err = pw.Close()
		if err != nil {
			errch <- err
			return
		}
		errch <- nil
	}()

	go func() {
		err := coding.Encode(dst, pr)
		if err != nil {
			errch <- err
			return
		}
		err = dst.Close()
		if err != nil {
			errch <- err
			return
		}
		err = pr.Close()
		if err != nil {
			errch <- err
			return
		}
		errch <- nil
	}()

	err = <-errch
	if err != nil {
		return err
	}
	err = <-errch // if the first one was nil, wait for the second one
	if err != nil {
		return err
	}

	return nil
}

func runZvmExecutable(zvme string, args ...string) error {
	toRun, err := os.CreateTemp("", "."+filepath.Base(zvme)+"*.temp.zvm")
	if err != nil {
		return err
	}
	defer func() {
		err = os.Remove(toRun.Name())
		if err != nil {
			fmt.Println("could not delete", toRun.Name(), err)
		}
	}()
	src, err := os.Open(zvme)
	if err != nil {
		panic(err)
	}

	if err = toRun.Chmod(0755); err != nil { // rwx r-x r-x
		return err
	}

	pr, pw := io.Pipe()

	errch := make(chan error)

	go func() {
		err = coding.Decode(pw, src)
		if err != nil {
			errch <- err
			return
		}
		err = src.Close()
		if err != nil {
			errch <- err
			return
		}
		err = pw.Close()
		if err != nil {
			errch <- err
			return
		}
		errch <- nil
	}()

	gzr, err := gzip.NewReader(pr)
	if err != nil {
		return err
	}

	go func() {
		_, err = io.Copy(toRun, gzr)
		if err != nil {
			errch <- err
			return
		}
		err = gzr.Close()
		if err != nil {
			errch <- err
			return
		}
		err = pr.Close()
		if err != nil {
			errch <- err
			return
		}
		errch <- nil
	}()

	err = <-errch
	if err != nil {
		return err
	}
	err = <-errch // if the first one was nil, wait for the second one
	if err != nil {
		return err
	}

	err = toRun.Sync()
	if err != nil {
		return err
	}
	err = toRun.Close()
	if err != nil {
		return err
	}
	cmd := exec.Command(toRun.Name(), args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	if err = cmd.Start(); err != nil {
		return err
	}
	err = cmd.Wait()
	return err
}
