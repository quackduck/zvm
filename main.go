package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/bearmini/bitstream-go"
	"io"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	encodeHaHa = []rune("zachZACH𝐳𝐚𝐜𝐡𝚣𝚊𝚌𝚑𝕫𝕒𝕔𝕙𝘇𝗮𝗰𝗵𝙯𝙖𝙘𝙝𝓏𝒶𝒸𝒽") //  zachZACH𝐳𝐚𝐜𝐡𝚣𝚊𝚌𝚑𝕫𝕒𝕔𝕙𝘇𝗮𝗰𝗵𝙯𝙖𝙘𝙝𝓏𝒶𝒸𝒽
	numOfBits = int(math.Log2(float64(len(encodeHaHa))))
)

func main() {
	if os.Args[1] == "--help" {
		fmt.Println(`Hey!!
Import commands like this: zvm import <file>
Run commands like this: zvm <zvm file name without .zvm)

Example session:
   cp $(which ls) ls
   zvm import ls
   zvm ls -l
`)
		return
	}
	rand.Seed(time.Now().UnixNano())
	if os.Args[1] == "import" {
		err := makeZvmExecutable(os.Args[2])
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	err := runZvmExecutable(os.Args[1] + ".zvm", os.Args[2:]...)
	if err != nil {
		fmt.Println(err)
	}
}

func makeZvmExecutable(exe string) error {
	dst, err := os.Create(exe + ".zvm")
	if err != nil {
		return err
	}
	src, err := os.Open(exe)
	if err != nil {
		return err
	}
	return encode(dst, src)
}

func runZvmExecutable(zvme string, args ...string) error {
	dst, err := os.Create(zvme + strconv.Itoa(rand.Int())+".temp-exe")
	//dst, err := os.CreateTemp("", "."+zvme +"*.temp.zvm")
	if err != nil {
		return err
	}
	//defer os.Remove(dst.Name())
	if err = dst.Chmod(0755); err != nil { // rwx r-x r-x
		return err
	}
	src, err := os.Open(zvme)
	if err != nil {
		return err
	}
	err = decode(dst, src)
	if err != nil {
		return err
	}
	return nil
	cmd := exec.Command("./"+dst.Name(), args...)
	//cmd := exec.Command(dst.Name(), args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	if err = cmd.Start(); err != nil {
		//fmt.Println("start error:")
		return err
	}
	//time.Sleep(2*time.Second)
	err = cmd.Wait()
	//e := err.(*exec.ExitError)
	//fmt.Println( e, e.String(), e.Sys(), e.Success(),e.Sys().(syscall.WaitStatus).TrapCause())
	return err
	//return cmd.Run()
}

// sliceByteLen slices the byte b such that the result has length len and starting bit start
func sliceByteLen(b byte, start int, len int) byte {
	return (b << start) >> byte(8-len)
}

func useHexEncode(dst io.Writer, src io.Reader) error {
	buf := make([]byte, 10*1024)
	replacer := strings.NewReplacer("0", "z", // 𝐳𝐚𝐜𝐡𝚣𝚊𝚌𝚑 𝕫𝕒𝕔𝕙𝘇𝗮𝗰𝗵𝙯𝙖𝙘𝙝𝓏𝒶𝒸𝒽
		"1", "a",
		"2", "c",
		"3", "h",
		"4", "Z",
		"5", "A",
		"6", "C",
		"7", "H",
		"8", "𝐳",
		"9", "𝐚",
		"a", "𝐜",
		"b", "𝚑",
		"c", "𝕫",
		"d", "𝕒",
		"e", "𝕔",
		"f", "𝕙",
		)
	//hexer := hex.NewEncoder()
	for {
		n, err := src.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		_, err = replacer.WriteString(dst, hex.EncodeToString(buf[:n]))
		if err != nil {
			return nil
		}
	}
	return nil
}

func useHexDecode(dst io.Writer, src io.Reader) error {
	buf := make([]byte, 10*1024)
	replacer := strings.NewReplacer( "z", "0",// 𝐳𝐚𝐜𝐡𝚣𝚊𝚌𝚑 𝕫𝕒𝕔𝕙𝘇𝗮𝗰𝗵𝙯𝙖𝙘𝙝𝓏𝒶𝒸𝒽
		 "a","1",
		 "c","2",
		 "h","3",
		"Z","4",
		 "A","5",
		"C","6",
		 "H","7",
		 "𝐳","8",
		 "𝐚","9",
		"𝐜","a",
		 "𝚑","b",
		 "𝕫","c",
		"𝕒","d",
		"𝕔","e",
		"𝕙","f",
	)
	//hexer := hex.NewEncoder()
	for {
		n, err := src.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		//_, err = replacer.WriteString(dst, hex.EncodeToString(buf[:n]))
		res, err := hex.DecodeString(replacer.Replace(string(buf[:n])))
		if err != nil {
			return nil
		}
		_, err = dst.Write(res)
		if err != nil {
			return nil
		}
	}
	return nil
}

func encode(dst io.Writer, src io.Reader) error {

	r := bitstream.NewReader(src, nil)

	//bs := bitStreamer{chunkLen: numOfBits, in: src}
	//err := bs.init()
	//if err != nil {
	//	panic(err)
	//}
	res := make([]byte, 0, 1024)
	for {
		chunk, err := r.ReadNBitsAsUint8(uint8(numOfBits))
		//chunk, err := bs.next()
		if err != nil {
			if err == io.EOF {
				_, err = dst.Write(res)
				//if err != nil {
				//	return err
				//}
				return err
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
	//bw := bitWriter{chunkLen: numOfBits, out: dst}
	//bw.init()

	br :=  bufio.NewReader(src)
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
				//fmt.Println(numOfBits)
				err := w.WriteNBitsOfUint8(uint8(numOfBits), byte(i))
				//err := bw.write(byte(i), numOfBits)
				if err != nil {
					return err
				}
			}
		}
	}
	return w.Flush()
	//return bw.flush()


	//w := bitstream.NewWriter(dst)
	////bw := bitWriter{chunkLen: numOfBits, out: dst}
	////bw.init()
	//buf := make([]byte, 10*1024)
	//prevBrokenBytes := make([]byte, 0, 5) // to hold last 5 bytes of each iteration so that there is no breaking on multi-byte unicode chars
	//for {
	//	n, err := src.Read(buf)
	//	if err != nil {
	//		if err == io.EOF {
	//			break
	//		}
	//		return err
	//	}
	//	var brokenBytes int
	//	if r, size := utf8.DecodeLastRune(buf[:n]); r == utf8.RuneError {
	//		brokenBytes = 1
	//		for ; r == utf8.RuneError ; brokenBytes++ {
	//			r, size = utf8.DecodeLastRune(buf[:n-brokenBytes])
	//		}
	//
	//		brokenBytes -= size
	//		if brokenBytes < 0 {
	//			brokenBytes = 0
	//		}
	//
	//		buf2 := make([]byte, 4)
	//		_ = utf8.EncodeRune(buf2, r)
	//		fmt.Println(string(r), buf2, buf[n-brokenBytes:n], buf[n-brokenBytes-size:n])
	//		// size is the len of the perfectly fine rune.
	//		// Consider scenario: <2 byte char> <garbage byte> <garbage byte>.
	//		// DecodeLastRune will only succeed when brokenBytes points before <2 byte char>, but that's not where the garbage starts.
	//	}
	//	fmt.Println(n-brokenBytes, n, brokenBytes)
	//
	//	prevBrokenBytesTemp := make([]byte, 0, brokenBytes)
	//	copy(prevBrokenBytesTemp, buf[n-brokenBytes:n]) // we'll add em broken bytes to the front of the next read
	//	buf = append(prevBrokenBytes, buf[:n-brokenBytes]...) // add last read's broken bytes to the front of this read
	//	prevBrokenBytes = prevBrokenBytesTemp
	//
	//	for _, c := range []rune(string(buf[:n])) {
	//		//fmt.Print(string(c))
	//		for i, char := range encodeHaHa {
	//			if c == char {
	//				//fmt.Println(numOfBits)
	//				err := w.WriteNBitsOfUint8(uint8(numOfBits), byte(i))
	//				//err := bw.write(byte(i), numOfBits)
	//				if err != nil {
	//					return err
	//				}
	//			}
	//		}
	//	}
	//}
	//return w.Flush()
	//return bw.flush()
}

type bitStreamer struct {
	// set these
	chunkLen int
	in       io.Reader

	// internal vars
	buf    []byte
	bitIdx int
	bufN   int
}

func (bs *bitStreamer) init() error {
	bs.buf = make([]byte, 16*1024)
	n, err := bs.in.Read(bs.buf)
	if err != nil {
		return err
	}
	bs.bufN = n
	return nil
}

func (bs *bitStreamer) next() (b byte, e error) {
	byteNum := bs.bitIdx / 8
	bitNum := bs.bitIdx % 8
	if byteNum >= bs.bufN { // need to read more?
		n, err := bs.in.Read(bs.buf)
		if err != nil {
			return 0, err
		}
		bs.bitIdx = bitNum
		byteNum = bs.bitIdx / 8
		bitNum = bs.bitIdx % 8
		bs.bufN = n
	}

	var result byte
	if bitNum+bs.chunkLen > 8 { // want to slice past current byte
		currByte := bs.buf[byteNum]
		didChange := false
		if byteNum+1 >= bs.bufN { // unlikely
			//fmt.Println("OMG OMG OMG OMG HELLO                                HELLO")
			didChange = true
			eh := make([]byte, 1)
			_, err := bs.in.Read(eh) // the actual data size doesn't change so we won't change n
			if err != nil {
				eh[0] = 0 // let it read from null byte (size can be inferred automatically at decoder (result has to be multiples of 8 bits))
				bs.bufN-- // next call should simpy exit so we make it as if there isn't any more data (which is actually already true)
			}
			if byteNum+1 >= len(bs.buf) {
				bs.buf = append(bs.buf, eh[0])
			} else {
				bs.buf[byteNum+1] = eh[0]
			}
			bs.bufN++
		}
		nextByte := bs.buf[byteNum+1]

		firstByte := sliceByteLen(currByte, bitNum, 8-bitNum)
		result = (firstByte << byte(bs.chunkLen+bitNum-8)) + sliceByteLen(nextByte, 0, bs.chunkLen+bitNum-8)
		if didChange {
			bs.bitIdx += bs.chunkLen - (8 - bitNum)
		}
	} else {
		result = sliceByteLen(bs.buf[byteNum], bitNum, bs.chunkLen)
	}
	bs.bitIdx += bs.chunkLen
	return result, nil
}

type bitWriter struct {
	chunkLen int
	out      io.Writer

	buf    []byte
	bitIdx int
}

func (bw *bitWriter) init() {
	bw.buf = make([]byte, 16*1024)
}

func (bw *bitWriter) write(b byte, bLen int) error {
	bitNum := bw.bitIdx % 8
	byteNum := bw.bitIdx / 8
	if byteNum >= len(bw.buf) {
		_, err := bw.out.Write(bw.buf)
		if err != nil {
			return err
		}
		bw.init()
		bw.bitIdx = 0
		bitNum = bw.bitIdx % 8
		byteNum = bw.bitIdx / 8
	}

	if 8-bitNum-bLen >= 0 {
		bw.buf[byteNum] = bw.buf[byteNum] + (b << (8 - bitNum - bLen))
	} else {
		bw.buf[byteNum] = bw.buf[byteNum] + sliceByteLen(b, 8-bLen, 8-bitNum)
		if len(bw.buf) <= byteNum+1 {
			_, err := bw.out.Write(bw.buf[:byteNum+1])
			if err != nil {
				return err
			}
			bw.init()
			bw.buf[0] = sliceByteLen(b, 8-bLen+8-bitNum, bLen+bitNum-8) << byte(8-bLen+8-bitNum)
			bw.bitIdx = 0
			byteNum = 0
			bitNum = 0
		} else {
			bw.buf[byteNum+1] = sliceByteLen(b, 8-bLen+8-bitNum, bLen+bitNum-8) << byte(8-bLen+8-bitNum)
		}
	}
	bw.bitIdx += bLen
	return nil
}

// call this only at the end
func (bw *bitWriter) flush() error {
	_, err := bw.out.Write(bw.buf[:bw.bitIdx/8])
	return err
}