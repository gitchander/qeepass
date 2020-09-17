package main

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/howeyc/gopass"
)

// https://gist.github.com/msmuenchen/9318327
// https://keepass.info/help/kb/kdbx_4.html#innerhdr
// https://github.com/keepassx/keepassx
// github.com/cixtor/kdbx
// github.com/tobischo/gokeepasslib

const (
	SIGNATURE1 uint32 = 0x9AA2D903
	SIGNATURE2 uint32 = 0xB54BFB67
)

const (
	COMPRESSION_NONE = 0
	COMPRESSION_GZIP = 1
)

// WORD  = 2 bytes, 16-bit unsigned integer
// DWORD = 4 bytes (double word), 32-bit unsigned integer
// QWORD = 8 bytes (quad word), 64-bit unsigned integer

// enum HeaderFieldID
// {
//     EndOfHeader = 0,
//     Comment = 1,
//     CipherID = 2,
//     CompressionFlags = 3,
//     MasterSeed = 4,
//     TransformSeed = 5,
//     TransformRounds = 6,
//     EncryptionIV = 7,
//     ProtectedStreamKey = 8,
//     StreamStartBytes = 9,
//     InnerRandomStreamID = 10
// };
// dynamic_header_type
const (
	BID_END                    = 0  // 5.1) bId=0: END entry, no more header entries after this
	BID_COMMENT                = 1  // 5.2) bId=1: COMMENT entry, unknown
	BID_CIPHER_ID              = 2  // 5.3) bId=2: CIPHERID, bData="31c1f2e6bf714350be5805216afc5aff" => outer encryption AES256, currently no others supported
	BID_COMPRESSION_FLAGS      = 3  // 5.4) bId=3: COMPRESSIONFLAGS, LE DWORD. 0=payload not compressed, 1=payload compressed with GZip
	BID_MASTER_SEED            = 4  // 5.5) bId=4: MASTERSEED, 32 BYTEs string. See further down for usage/purpose. Length MUST be checked.
	BID_TRANSFORM_SEED         = 5  // 5.6) bId=5: TRANSFORMSEED, variable length BYTE string. See further down for usage/purpose.
	BID_TRANSFORM_ROUNDS       = 6  // 5.7) bId=6: TRANSFORMROUNDS, LE QWORD. See further down for usage/purpose.
	BID_ENCRYPTION_IV          = 7  // 5.8) bId=7: ENCRYPTIONIV, variable length BYTE string. See further down for usage/purpose.
	BID_PROTECTED_STREAM_KEY   = 8  // 5.9) bId=8: PROTECTEDSTREAMKEY, variable length BYTE string. See further down for usage/purpose.
	BID_STREAM_START_BYTES     = 9  // 5.10) bId=9: STREAMSTARTBYTES, variable length BYTE string. See further down for usage/purpose.
	BID_INNER_RANDOM_STREAM_ID = 10 // 5.11) bId=10: INNERRANDOMSTREAMID, LE DWORD. Inner stream encryption type, 0=>none, 1=>Arc4Variant, 2=>Salsa20
)

type FileSignatures struct {
	Signature1   uint32
	Signature2   uint32
	VersionMinor uint16
	VersionMajor uint16
}

type FileHeaders struct {
	Comment             []byte // FieldID: 1
	CipherID            []byte // FieldID: 2
	CompressionFlags    uint32 // FieldID: 3
	MasterSeed          []byte // FieldID: 4
	TransformSeed       []byte // FieldID: 5 (KDBX 3.1)
	TransformRounds     uint64 // FieldID: 6 (KDBX 3.1)
	EncryptionIV        []byte // FieldID: 7
	ProtectedStreamKey  []byte // FieldID: 8 (KDBX 3.1)
	StreamStartBytes    []byte // FieldID: 9 (KDBX 3.1)
	InnerRandomStreamID uint32 // FieldID: 10 (KDBX 3.1)
}

type DynamicHeader struct {
	BID  byte
	Data []byte
}

func main() {

	if len(os.Args) < 2 {
		log.Fatal("need an argument")
	}
	filename := os.Args[1]

	fmt.Print("Enter master password ")
	secret, err := gopass.GetPasswd()
	checkError(err)

	err = readDatabase(filename, secret)
	checkError(err)
}

func readSignatures(data []byte, fsig *FileSignatures) (rest []byte, err error) {

	if len(data) < 12 {
		return data, ErrInsuffDataLen
	}

	signature1 := binary.LittleEndian.Uint32(data[0:4])
	if signature1 != SIGNATURE1 {
		return nil, errors.New("Not a KeePass database.")
	}

	signature2 := binary.LittleEndian.Uint32(data[4:8])
	if signature2 != SIGNATURE2 {
		return nil, errors.New("Not a KeePass database.")
	}

	fileVersionMinor := binary.LittleEndian.Uint16(data[8:10])
	fileVersionMajor := binary.LittleEndian.Uint16(data[10:12])

	data = data[12:]

	*fsig = FileSignatures{
		Signature1:   signature1,
		Signature2:   signature2,
		VersionMinor: fileVersionMinor,
		VersionMajor: fileVersionMajor,
	}

	return data, nil
}

func readFileHeaders(data []byte, fh *FileHeaders) (rest []byte, err error) {
	for {
		var dh DynamicHeader
		rest, err := readDynamicHeader(data, &dh)
		if err != nil {
			return rest, err
		}
		data = rest

		//printDynamicHeader(&dh)

		switch dh.BID {
		case BID_END:
			return data, nil
		case BID_COMMENT:
			fh.Comment = cloneBytes(dh.Data)
		case BID_CIPHER_ID:
			fh.CipherID = cloneBytes(dh.Data)
		case BID_COMPRESSION_FLAGS:
			{
				u := binary.LittleEndian.Uint32(dh.Data)
				fh.CompressionFlags = u
			}
		case BID_MASTER_SEED:
			fh.MasterSeed = cloneBytes(dh.Data)
		case BID_TRANSFORM_SEED:
			fh.TransformSeed = cloneBytes(dh.Data)
		case BID_TRANSFORM_ROUNDS:
			{
				u := binary.LittleEndian.Uint64(dh.Data)
				fh.TransformRounds = u
			}
		case BID_ENCRYPTION_IV:
			fh.EncryptionIV = cloneBytes(dh.Data)
		case BID_PROTECTED_STREAM_KEY:
			fh.ProtectedStreamKey = cloneBytes(dh.Data)
		case BID_STREAM_START_BYTES:
			fh.StreamStartBytes = cloneBytes(dh.Data)
		case BID_INNER_RANDOM_STREAM_ID:
			{
				u := binary.LittleEndian.Uint32(dh.Data)
				fh.InnerRandomStreamID = u
			}
		}
	}
}

func readDynamicHeader(data []byte, p *DynamicHeader) ([]byte, error) {

	if len(data) < 3 {
		return data, ErrInsuffDataLen
	}

	//-------------------------------
	// Read BID:
	p.BID = data[0]
	data = data[1:]
	//-------------------------------
	// Read Size:
	size := binary.LittleEndian.Uint16(data[0:2])
	data = data[2:]
	//-------------------------------
	// Read Data:
	if len(data) < int(size) {
		return data, ErrInsuffDataLen
	}
	p.Data = cloneBytes(data[:size])
	data = data[size:]
	//-------------------------------

	return data, nil
}

func printDynamicHeader(dh *DynamicHeader) {

	var sval string
	switch len(dh.Data) {
	case 2:
		u := binary.LittleEndian.Uint16(dh.Data)
		sval = "val:" + strconv.FormatUint(uint64(u), 10)
	case 4:
		u := binary.LittleEndian.Uint32(dh.Data)
		sval = "val:" + strconv.FormatUint(uint64(u), 10)
	case 8:
		u := binary.LittleEndian.Uint64(dh.Data)
		sval = "val:" + strconv.FormatUint(u, 10)
	}

	fmt.Printf("bid:%d, data:[%x] %s\n", dh.BID, dh.Data, sval)
}

func readDatabase(filename string, secret []byte) error {

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var fsig FileSignatures
	data, err = readSignatures(data, &fsig)
	if err != nil {
		return err
	}

	var fh FileHeaders
	data, err = readFileHeaders(data, &fh)
	if err != nil {
		return err
	}

	encrypted := data

	key, err := buildMasterKey(&fh, secret)
	if err != nil {
		return err
	}

	b, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	bm := cipher.NewCBCDecrypter(b, fh.EncryptionIV)

	decryptedContent := make([]byte, len(encrypted))
	bm.CryptBlocks(decryptedContent, encrypted)

	startBytes := fh.StreamStartBytes
	startBytesHave := decryptedContent[:len(startBytes)]
	if !bytes.Equal(startBytes, startBytesHave) {
		return errors.New("kdbx.content: invalid auth or corrupt database")
	}

	decryptedContent = decryptedContent[len(startBytes):]

	if fsig.VersionMajor != 4 {
		reader := bytes.NewReader(decryptedContent)
		decryptedContent, err = decomposeContentBlocks31(reader)
		if err != nil {
			return err
		}
	}

	if fh.CompressionFlags == COMPRESSION_GZIP {
		reader := bytes.NewReader(decryptedContent)
		z, err := gzip.NewReader(reader)
		if err != nil {
			return err
		}
		decryptedContent, err = ioutil.ReadAll(z)
		if err != nil {
			return err
		}
	}

	fmt.Println(string(decryptedContent))

	return nil
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var ErrInsuffDataLen = errors.New("insufficient data length")

func cloneBytes(a []byte) []byte {
	b := make([]byte, len(a))
	copy(b, a)
	return b
}

func sum256N(data []byte, n int) []byte {
	for i := 0; i < n; i++ {
		h := sha256.Sum256(data)
		data = h[:]
	}
	return data
}

func buildMasterKey(fh *FileHeaders, secret []byte) ([]byte, error) {

	key := sum256N(secret, 2)

	h := sha256.New()
	h.Write(fh.MasterSeed)

	block, err := aes.NewCipher(fh.TransformSeed)
	if err != nil {
		return nil, err
	}

	tkey := make([]byte, len(key))
	copy(tkey, key)

	rounds := fh.TransformRounds
	const k = 16
	for i := uint64(0); i < rounds; i++ {
		block.Encrypt(tkey[:k], tkey[:k])
		block.Encrypt(tkey[k:], tkey[k:])
	}

	tmp := sha256.Sum256(tkey)
	tkey = tmp[:]

	tkey = append(fh.MasterSeed, tkey...)
	hsh := sha256.Sum256(tkey)
	tkey = hsh[:]

	return tkey, nil
}

func printJSON(prefix string, v interface{}) {
	data, err := json.MarshalIndent(v, "", "\t")
	checkError(err)
	fmt.Println(prefix, string(data))
}

// decomposeContentBlocks31 decodes the content data block by block (Kdbx v3.1)
// Used to extract data blocks from the entire content
func decomposeContentBlocks31(r io.Reader) ([]byte, error) {
	var contentData []byte
	// Get all the content
	content, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	offset := uint32(0)
	for {
		var hash [32]byte
		var length uint32
		var data []byte

		// Skipping Index, uint32
		offset = offset + 4

		copy(hash[:], content[offset:offset+32])
		offset = offset + 32

		length = binary.LittleEndian.Uint32(content[offset : offset+4])
		offset = offset + 4

		if length > 0 {
			data = make([]byte, length)
			copy(data, content[offset:offset+length])
			offset = offset + length

			// Add to decoded blocks
			contentData = append(contentData, data...)
		} else {
			break
		}
	}
	return contentData, nil
}
