package mimime

import (
    "crypto/md5"
    "errors"
    "fmt"
    "io"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "strconv"
    "strings"
    "sync"
    "github.com/sellleon/mimime/fsm"
)

type FileUnit string
type Option int
type ResizeFlag int
type OptionRegistration (func(r *RequestOptions, arg string) error)

const (
    B  FileUnit = "B"
    KB FileUnit = "KB"
    MB FileUnit = "MB"
    GB FileUnit = "GB"
)

var RegisteredOptions map[string]OptionRegistration
var FileLocks map[string]*sync.Mutex
var FileLocksLock sync.Mutex

const Name string = "mimime"

var (
    HomePath      string
    CachePath     string
    CacheOrigPath string
)

const (
    FileSizeOption    Option = iota
    SslOption         Option = iota
    ForceReloadOption Option = iota
    GrayScaleOption   Option = iota
    QualityOption     Option = iota
    ResizeOption      Option = iota
)

const (
    RFLeft  ResizeFlag = iota
    RFRight ResizeFlag = iota
    RFBoth  ResizeFlag = iota
    RFPerc  ResizeFlag = iota
)

const (
    StartState fsm.FsmState = iota
    LeftParseState fsm.FsmState = iota
    NoLeftParseState fsm.FsmState = iota
    LeftOnlyParseState fsm.FsmState = iota
    RightOnlyParseState fsm.FsmState = iota
    BothParseState fsm.FsmState = iota
)

const (
    LeftParseInput fsm.FsmInput = iota
    NoLeftParseInput fsm.FsmInput = iota
    RightParseInput fsm.FsmInput = iota
    NoRightParseInput fsm.FsmInput = iota
)

var ParseTransitions fsm.FsmTrans
var AcceptingParseStates []fsm.FsmState

var (
    DefaultFileSize FileSize = FileSize{50, KB}
)

type FileSize struct {
    value float64
    unit  FileUnit
}

type ImgSize struct {
    rf     ResizeFlag
    width  int64
    height int64
    perc   float64
}

type RequestOptions struct {
    setOpts map[Option]bool
    fs      FileSize
    qual    float64
    re      ImgSize
}

type Request struct {
    imgUrl  string
    imgId   string
    reqOpts RequestOptions
}

func (fs FileSize) String() string {
    return fmt.Sprintf("%f%s", fs.value, fs.unit)
}

func (r *Request) GImgId() string {
    if r.imgId == "" {
        r.imgId = fmt.Sprintf("%x", md5.Sum([]byte(r.imgUrl)))
    }
    return r.imgId
}

func ResizeFlagFromState(s fsm.FsmState) (ResizeFlag, error) {
    switch s {
    case LeftOnlyParseState:
        return RFLeft, nil
    case RightOnlyParseState:
        return RFRight, nil
    case BothParseState:
        return RFBoth, nil
    }
    return -1, errors.New("Invalid size option given.")
}

func (r *Request) OrigPath() string {
    return filepath.Join(CacheOrigPath, r.GImgId())
}

func init() {
    HomePath = os.Getenv("HOME")
    CachePath = filepath.Join(HomePath, ".cache", Name)
    CacheOrigPath = filepath.Join(CachePath, "orig")

    err := os.MkdirAll(HomePath, os.ModePerm)
    if err != nil {
        fmt.Println(err)
    }
    err = os.MkdirAll(CachePath, os.ModePerm)
    if err != nil {
        fmt.Println(err)
    }
    err = os.MkdirAll(CacheOrigPath, os.ModePerm)
    if err != nil {
        fmt.Println(err)
    }

    startStateTrans := make(map[fsm.FsmInput]fsm.FsmState)
    startStateTrans[LeftParseInput] = LeftParseState
    startStateTrans[NoLeftParseInput] = NoLeftParseState
    leftParseStateTrans := make(map[fsm.FsmInput]fsm.FsmState)
    leftParseStateTrans[RightParseInput] = BothParseState
    leftParseStateTrans[NoRightParseInput] = LeftOnlyParseState
    noLeftParseStateTrans := make(map[fsm.FsmInput]fsm.FsmState)
    noLeftParseStateTrans[RightParseInput] = RightOnlyParseState
    ParseTransitions = make(fsm.FsmTrans)
    ParseTransitions[StartState] = startStateTrans
    ParseTransitions[LeftParseState] = leftParseStateTrans
    ParseTransitions[NoLeftParseState] = noLeftParseStateTrans

    AcceptingParseStates = []fsm.FsmState{
                                LeftOnlyParseState,
                                RightOnlyParseState,
                                BothParseState}

    RegisteredOptions = make(map[string]OptionRegistration)
    FileLocks = make(map[string]*sync.Mutex)
    addOption([]string{"s", "-size", "-filesize"}, RegisterFileSizeOption)
    addOption([]string{"p", "-ssl"}, GenToggleOptionRegistration(SslOption))
    addOption([]string{"f", "-force"}, GenToggleOptionRegistration(ForceReloadOption))
    addOption([]string{"g", "-gray"}, GenToggleOptionRegistration(GrayScaleOption))
    addOption([]string{"q", "-quality"}, RegisterQualityOption)
    addOption([]string{"r", "-resize"}, RegisterResizeOption)
}

func LogErr(w io.Writer, err error) {
    fmt.Fprintf(w, "Fatal error: %s", err)
}

func addOption(prefixes []string, or OptionRegistration) {
    for _, el := range prefixes {
        RegisteredOptions[el] = or
    }
}

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil {
        return true, nil
    }
    if os.IsNotExist(err) {
        return false, nil
    }
    return true, err
}

func ParseFileUnit(unparsedFileUnit string) (*FileUnit, error) {
    var fu FileUnit
    switch strings.ToLower(unparsedFileUnit) {
    case "b":
        fu = B
    case "kb":
        fu = KB
    case "mb":
        fu = MB
    case "gb":
        fu = GB
    default:
        fu = KB
    }
    return &fu, nil
}

func ParseFileSize(unparsedFileSize string) (*FileSize, error) {
    splitIndex := len(unparsedFileSize)
    for index, character := range unparsedFileSize {
        if !strings.Contains("1234567890.", string(character)) {
            splitIndex = index
            break
        }
    }
    size, err := strconv.ParseFloat(unparsedFileSize[:splitIndex], 64)
    if err != nil {
        return nil, err
    }
    fu, err := ParseFileUnit(unparsedFileSize[splitIndex:])
    if err != nil {
        return nil, err
    }
    return &FileSize{size, *fu}, nil
}

func RegisterQualityOption(r *RequestOptions, arg string) error {
    quality, err := strconv.ParseFloat(arg, 64)
    if err != nil {
        return err
    }
    r.setOpts[QualityOption] = true
    r.qual = quality
    return nil
}

func RegisterResizeOption(r *RequestOptions, arg string) error {
    if strings.Contains(arg, "x") {
        sizes := strings.Split(arg, "x")
        if len(sizes) != 2 {
            return errors.New(fmt.Sprintf("Invalid amount of size parameters: %s", arg))
        }
        var width  int64
        var height int64
        var err    error

        parseFsm := fsm.Initialize(ParseTransitions, StartState, AcceptingParseStates)

        if sizes[0] == "" {
            parseFsm.Advance(NoLeftParseInput)
        } else {
            width, err = strconv.ParseInt(sizes[0], 10, 64)
            if err != nil {
                return err
            }
            parseFsm.Advance(LeftParseInput)
        }

        if sizes[1] == "" {
            parseFsm.Advance(NoRightParseInput)
        } else {
            height, err = strconv.ParseInt(sizes[1], 10, 64)
            if err != nil {
                return err
            }
            parseFsm.Advance(RightParseInput)
        }

        finalState, err := parseFsm.Finalize()
        if err != nil {
            return errors.New("Invalid size option given.")
        }

        flag, err := ResizeFlagFromState(finalState)
        if err != nil {
            return err
        }
        r.setOpts[ResizeOption] = true
        r.re = ImgSize{rf: flag, width: width, height: height}
        return nil
    }
    perc, err := strconv.ParseFloat(arg, 64)
    if err != nil {
        return err
    }
    r.setOpts[ResizeOption] = true
    r.re = ImgSize{rf: RFPerc, perc: perc}
    return nil
}

func RegisterFileSizeOption(r *RequestOptions, arg string) error {
    fs, err := ParseFileSize(arg)
    if err != nil {
        return err
    }
    r.setOpts[FileSizeOption] = true
    r.fs = *fs
    return nil
}

func GenToggleOptionRegistration(opt Option) OptionRegistration {
    return func(r *RequestOptions, arg string) error {
        r.setOpts[opt] = true
        return nil
    }
}

func RegisterOption(ro *RequestOptions, arg string) error {
    for pre, or := range RegisteredOptions {
        if strings.HasPrefix(arg, pre) {
            return or(ro, strings.TrimPrefix(arg, pre))
        }
    }
    errMsg := fmt.Sprintf("Unknown option: %s", arg)
    return errors.New(errMsg)
}

func ParseRequestOptions(unparsedOptions []string) (*RequestOptions, error) {
    setOpts := make(map[Option]bool)
    ro := &RequestOptions{setOpts: setOpts}
    for _, el := range unparsedOptions {
        err := RegisterOption(ro, el)
        if err != nil {
            return nil, err
        }
    }
    return ro, nil
}

func ParseUrl(unparsedUrl string) (string, bool, error) {
    if strings.HasPrefix(unparsedUrl, "http:/") {
        return unparsedUrl[6:], false, nil
    }
    if strings.HasPrefix(unparsedUrl, "https:/") {
        return unparsedUrl[7:], true, nil
    }
    return unparsedUrl, false, nil
}

func ParseRequest(path string) (*Request, error) {
    urlSplit := strings.SplitN(path, "/u", 2)
    var imgUrl string
    var options []string
    var sslFlag bool
    var err error
    if len(urlSplit) > 1 {
        imgUrl, sslFlag, err = ParseUrl(urlSplit[1])
        if err != nil {
            return nil, err
        }
        options = strings.Split(urlSplit[0], "/")
    } else {
        imgUrl, sslFlag, err = ParseUrl(urlSplit[0][1:])
        if err != nil {
            return nil, err
        }
        options = []string{""}
    }
    reqOpts, err := ParseRequestOptions(options[1:])
    if err != nil {
        return nil, err
    }
    if sslFlag {
        reqOpts.setOpts[SslOption] = true;
    }
    return &Request{imgUrl, "", *reqOpts}, nil
}

func LogRequest(req *Request) {
    fmt.Printf("Requesting image %s\n", req.imgUrl)

    sslString := "disabled"
    if req.reqOpts.setOpts[SslOption] {
        sslString = "enabled"
    }
    fmt.Printf("SSL option is %s\n", sslString)
    forceString := "disabled"
    if req.reqOpts.setOpts[ForceReloadOption] {
        forceString = "enabled"
    }
    fmt.Printf("Force reload option is %s\n", forceString)

    if req.reqOpts.setOpts[FileSizeOption] {
        fmt.Printf("Option FileSize is set with value: %s\n", req.reqOpts.fs)
    } else {
        fmt.Printf("No FileSize set using default value: %s\n", req.reqOpts.fs)
    }
}

func WriteRequestResponse(req *Request, path string) error {
    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer f.Close()

    protocol := "http"
    if req.reqOpts.setOpts[SslOption] {
        protocol = "https"
    }

    response, err := http.Get(protocol + "://" + req.imgUrl)
    if err != nil {
        return err
    }
    defer response.Body.Close()

    _, err = io.Copy(f, response.Body)
    if err != nil {
        return err
    }

    return nil
}

func CreateOriginalFile(req *Request) error {
    path := filepath.Join(CacheOrigPath, req.GImgId())
    CoLock(req.GImgId())
    defer CoUnlock(req.GImgId())

    if !req.reqOpts.setOpts[ForceReloadOption] {
        ex, err := exists(path)
        if err != nil {
            return err
        }
        if ex {
            return nil
        }
    }

    return WriteRequestResponse(req, path)
}

func CoLock(key string) {
    FileLocksLock.Lock()
    lock, ok := FileLocks[key]
    if !ok {
        lock = &sync.Mutex{}
        FileLocks[key] = lock
    }
    lock.Lock()
    FileLocksLock.Unlock()
}

func CoUnlock(key string) {
    FileLocksLock.Lock()
    FileLocks[key].Unlock()
    FileLocksLock.Unlock()
}

func Minify(req *Request) (*exec.Cmd, error) {
    fn := "convert"
    args := []string{}
    for key, value := range req.reqOpts.setOpts {
        if value {
            switch key {
            case FileSizeOption:
                args = append(args, "-define", "jpeg:extent="+req.reqOpts.fs.String())
            case GrayScaleOption:
                args = append(args, "-colorspace", "Gray")
            case QualityOption:
                args = append(
                    args,
                    "-quality",
                    fmt.Sprintf("%.6f%%", req.reqOpts.qual))
            case ResizeOption:
                re := req.reqOpts.re
                switch re.rf {
                case RFLeft:
                    args = append(args, "-resize", fmt.Sprintf("%dx", re.width))
                case RFRight:
                    args = append(args, "-resize", fmt.Sprintf("x%d", re.height))
                case RFBoth:
                    args = append(
                        args,
                        "-resize",
                        fmt.Sprintf("%dx%d", re.width, re.height))
                case RFPerc:
                    args = append(
                        args,
                        "-resize",
                        fmt.Sprintf("%.6f%%", re.perc))
                default:
                    return nil, errors.New("Invalid resize flag.")
                }
            }
        }
    }
    args = append(args, req.OrigPath())
    args = append(args, "jpeg:-")
    return exec.Command(fn, args...), nil
}

func handler(w http.ResponseWriter, r *http.Request) {
    req, err := ParseRequest(r.URL.Path)
    if err != nil {
        LogErr(w, err)
        return
    }
    go LogRequest(req)

    err = CreateOriginalFile(req)
    if err != nil {
        LogErr(w, err)
        return
    }

    /*
    * <- gather info
    */

    cmd, err := Minify(req)
    if err != nil {
        LogErr(w, err)
        return
    }

    stdoutPipe, err := cmd.StdoutPipe()
    if err != nil {
        LogErr(w, err)
        return
    }

    err = cmd.Start()
    if err != nil {
        LogErr(w, err)
        return
    }


    _, err = io.Copy(w, stdoutPipe)
    if err != nil {
        LogErr(w, err)
        return
    }
    cmd.Wait()
}

func RunServer() {
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}
