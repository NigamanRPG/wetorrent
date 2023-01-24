package main

import (
	"log"
	"net/http"
	"strings"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	//"fyne.io/fyne/v2/canvas"
	"net/url"
	"time"	
    "encoding/json"
    "fmt"
    "os"
	"github.com/gorilla/websocket"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	//"github.com/anacrolix/torrent/storage"
)

var upgrader = websocket.Upgrader{}




var mainapp fyne.App

var MainTorrent string//magnet
var MainFile string//filepath
var AppIsClosing bool
func main() {
	mainapp= app.New()
	mainapp.Settings().SetTheme(&myTheme{})
	mainapp.SetIcon(resourceAppiconPng)
	mainwin := mainapp.NewWindow("wetorrent")
	mainwin.Resize(fyne.NewSize(400, 710))
	go startWebsocket()
	go startServer()
	AppIsClosing=false
	go initmainclient()
	LoadSettings()
	//time.Sleep(10*time.Second)
	//go SetMainTorrent("magnet:?xt=urn:btih:D7A46713EAEE18C746B3254B7D1492A50FD9D6CE&dn=The+Matrix+%281999%29+%5B1080p%5D+%5BYTS.MX%5D&tr=udp%3A%2F%2Fglotorrents.pw%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.openbittorrent.com%3A80&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Fp4p.arenabg.ch%3A1337&tr=udp%3A%2F%2Ftracker.internetwarriors.net%3A1337")
	//SetMainFile("The Matrix (1999) [1080p]/The.Matrix.1999.1080p.BrRip.x264.YIFY.mp4")
	//go addtorrent("magnet:?xt=urn:btih:D7A46713EAEE18C746B3254B7D1492A50FD9D6CE&dn=The+Matrix+%281999%29+%5B1080p%5D+%5BYTS.MX%5D&tr=udp%3A%2F%2Fglotorrents.pw%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.openbittorrent.com%3A80&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Fp4p.arenabg.ch%3A1337&tr=udp%3A%2F%2Ftracker.internetwarriors.net%3A1337")
	//w.SetContent(widget.NewLabel("Wetorrent is live ..."))
	tabs := container.NewAppTabs(
		container.NewTabItem("Home",  homeScreen(mainwin)),
		//container.NewTabItem("Settings",  settingsScreen(myWindow)),	
	)
		
	tabs.SetTabLocation(container.TabLocationTop)
		
	mainwin.SetContent(tabs)	
	
	
	mainwin.ShowAndRun()
	AppIsClosing=true

}

func homeScreen(win fyne.Window) fyne.CanvasObject {
	data := binding.BindStringList(
		//&[]string{"Item 1", "Item 2", "Item 3"},
		&[]string{},
	)

	list := widget.NewListWithData(data,
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*widget.Label).Bind(i.(binding.String))
		})

	add := widget.NewButton("Open New Webapp Tab", func() {
		//val := fmt.Sprintf("Item %d", data.Length()+1)
		//data.Append(val)
		openNewWebappTab()
		
	})
	return container.NewBorder( add,nil, nil, nil, list)	
}

func openNewWebappTab(){
	u, err := url.Parse("http://localhost:8080/wetorrent/wetorrent.html")
	_=err
	mainapp.OpenURL(u)  //

}

func startServer(){
	openNewWebappTab()


	
	//
	fs := http.FileServer(http.Dir("./Webapp"))
	http.Handle("/", http.StripPrefix("/", fs))

	fmt.Println(http.ListenAndServe(":8080", nil))

}

func startWebsocket(){
    http.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
        // Upgrade upgrades the HTTP server connection to the WebSocket protocol.
        conn, err := upgrader.Upgrade(w, r, nil)
        if err != nil {
            log.Print("upgrade failed: ", err)
            return
        }
        defer conn.Close()

        // Continuosly read and write message
        for {
            mt, message, err := conn.ReadMessage()
            if err != nil {
                log.Println("read failed:", err)
                //break
		//mainapp.Quit()
            }
		messagestring:=string(message)
		messageArr := strings.Split(messagestring, "*")
		log.Println("got:", messagestring)
		
		returnmessagestring:=runCmd(messageArr)//[]byte("return message")
            err = conn.WriteMessage(mt,[]byte(returnmessagestring))
            if err != nil {
                log.Println("write failed:", err)
                //break
            }
        }
    })


}

func runCmd(messageArr []string) string{
	if len(messageArr)==0{
		fmt.Println("Unkown command")
	}
	   switch messageArr[0] {
	    case "SETMAINTORRENT":
		fmt.Println("got cmd SYNCMAINTORRENT")
		//
		go SetMainTorrent(messageArr[1])
		if len(messageArr)>2{
			SetMainFile(messageArr[2])
		}
	    case "SETMAINFILE":
		fmt.Println("got cmd SYNCMAINFILE")
		//
		SetMainFile(messageArr[1])
	    case "ADDSAVEDITEM":
		fmt.Println("got cmd ADDSAVEDITEM")
		//
		AddSavedItem(messageArr[1],messageArr[2])
	    case "REMOVESAVEDITEM":
		fmt.Println("got cmd REMOVESAVEDITEM")
		//
		RemoveSavedItem(messageArr[1])
	    case "REQUESTTORRENTINFO":
		if len(messageArr)>1{
		return getTorrentInfoResponse(messageArr[1])
		}
	   case "REQUESTISSAVEDITEM":
		return getIsSavedItemResponse(messageArr[1])
	    default:
		fmt.Println("Unkown command")
	    }
	
 	return "return message"
}

var mainclient * torrent.Client

func initmainclient() {
	
	cfg := torrent.NewDefaultClientConfig()
	cfg.Seed = true
	cfg.DataDir="Webapp/wetorrent/torrents"//****************
	//cfg.NoDHT = true
	//cfg.DisableTCP = true
	//cfg.DisableUTP = true
	cfg.DisableAggressiveUpload = false
	cfg.DisableWebtorrent = false
	cfg.DisableWebseeds = false
	var err error
	mainclient, err = torrent.NewClient(cfg)
	if err != nil {
		log.Print("new torrent client: %w", err)
		return //fmt.Errorf("new torrent client: %w", err)
	}
	log.Print("new torrent client INITIATED")
	for {
		if AppIsClosing {
			log.Print("closing mainclient")
			mainclient.Close()
		}
		time.Sleep(1 * time.Second)
	}
	//
}
func SetMainFile(tmpfilepath string){
	MainFile=tmpfilepath
	//Prioritize(MainTorrent,tmpfilepath)
}
func SetMainTorrent(magnet string){
	if (!IsMainTorrent(magnet))&&(!IsSavedItemWithMagnet(magnet)){
		MainTorrent=magnet
			for {
				
				if (mainclient!=nil)&&(!AppIsClosing){
					break
				}
				time.Sleep(1 * time.Second)
			}
		addtorrent(magnet)
	}
}
func addtorrent(tmpmagneturi string) {
/*
	tmpmagnet,perr:=metainfo.ParseMagnetUri(tmpmagneturi)
	//_=perr
	if perr != nil {
		log.Print("new torrent parsing error: %w", perr)
		//return //fmt.Errorf("new torrent client: %w", err)
	}
*/
	t, err := mainclient.AddMagnet(tmpmagneturi)
	if err != nil {
		log.Print("new torrent error: %w", err)
		//return //fmt.Errorf("new torrent client: %w", err)
	}
/*
	t,ok:=mainclient.Torrent(tmpmagnet.InfoHash)
	//_=ok
	if (!ok) {
		log.Print("new torrent error ")
		return
		//return //fmt.Errorf("new torrent client: %w", err)
	}
*/
//
/*
	mms := storage.NewMMap("Webapp/wetorrent/torrents")
	defer mms.Close()
	tspec,perr:=torrent.TorrentSpecFromMagnetUri(tmpmagneturi)
 	_=perr
	log.Printf("torrent spec",tspec)
	tspec.Storage=mms//:   mms
	log.Printf("torrent spec",tspec)
	t, new, err := mainclient.AddTorrentSpec(tspec)
	_=new
	_=err

	if err != nil {
		log.Print("new torrent error: %w", err)
		//return //fmt.Errorf("new torrent client: %w", err)
	}
*/
	//
	//t, _ := mainclient.AddMagnet(tmpmagneturi)
	
	<-t.GotInfo()
	log.Printf("added magnet %s\n",tmpmagneturi)
	//t.DownloadAll()
	//mainclient.WaitAll()
	//selectedfilepath:="[TorrentCouch.com].Tom.Clancys.Jack.Ryan.S01.Complete.720p.WEB-DL.x264.[4.3GB].[MP4].[Season.1.Full]/[TorrentCouch.com].Tom.Clancys.Jack.Ryan.S01E04.720p.WEB-DL.x264.mp4"
	//Prioritize(tmpmagneturi,selectedfilepath)
	//
	for  {
		Prioritize(tmpmagneturi,MainFile)
		//DisplayTorrentInfo(tmpmagneturi)
		if (!IsSavedItemWithMagnet(tmpmagneturi))&&(!IsMainTorrent(tmpmagneturi)){
			log.Println("Torrent removed",tmpmagneturi)
			t.Drop()
			return
		}
		time.Sleep(8 * time.Second)
	}

	time.Sleep(1 * time.Second)
	log.Println("Torrent downloaded")
}
/*
func DisplayTorrentInfo(tmpmagneturi string){
		tmpmagnet,perr:=metainfo.ParseMagnetUri(tmpmagneturi)
		_=perr
		t,ok:=mainclient.Torrent(tmpmagnet.InfoHash)
		_=ok

		files:=t.Files()
			for i, filei := range files {
				fmt.Printf("**file %d path %s progress %d %% \n", i,filei.Path(), filei.BytesCompleted()*100/filei.Length())
			}
		fmt.Printf("***\n")

}*/
func getIsSavedItemResponse(tmpitempath string)string{
	var tmpreturnstring="ISSAVEDITEM*"+tmpitempath
	
	if IsSavedItem(tmpitempath){
		tmpreturnstring+="*TRUE"
	} else {
		tmpreturnstring+="*FALSE"
	}

	return tmpreturnstring

}
func getTorrentInfoResponse(tmpmagneturi string)string{
	
		fmt.Printf("REQUESTTORRENTINFO %s \n",tmpmagneturi)
		var tmpreturnstring="TORRENTINFO"
		tmpmagnet,perr:=metainfo.ParseMagnetUri(tmpmagneturi)
		_=perr
		if perr!=nil{
			return ""
		}
		t,ok:=mainclient.Torrent(tmpmagnet.InfoHash)
		_=ok
		if !ok{
			return ""
		}
		if t==nil{
			return ""
		}
		if t.Info()==nil {
			return ""
		}
		files:=t.Files()
		if files==nil{
			return ""
		}
		tmpreturnstring+="*"+tmpmagneturi
		tmpreturnstring+="*"+"TORRENTNAME"
		tmpreturnstring+="*"+fmt.Sprintf("%d",len(t.PeerConns()))//"333"//nbpeers
		//tmpreturnstring+="*"+fmt.Sprintf("%d",len(files))
		
			for _, filei := range files {
				//fmt.Printf("**file %d path %s progress %d %% \n", i,filei.Path(), filei.BytesCompleted()*100/filei.Length())
				tmpreturnstring+="*"+fmt.Sprintf("%s*%d",filei.Path(), filei.BytesCompleted()*100/filei.Length())
			}
		fmt.Printf("*** %s\n",tmpreturnstring)
		return tmpreturnstring
}

func Prioritize(tmpmagneturi string, filepath string){
		tmpmagnet,perr:=metainfo.ParseMagnetUri(tmpmagneturi)
		_=perr
		t,ok:=mainclient.Torrent(tmpmagnet.InfoHash)
		_=ok
		if !ok {
			return
		}

		files:=t.Files()
			for _, filei := range files {
				//fmt.Printf("**file %d path %s progress %d %% \n", i,filei.Path(), filei.BytesCompleted()*100/filei.Length())
				if filepath==filei.Path() {
					filei.SetPriority(torrent.PiecePriorityNormal)
				} else {
					filei.SetPriority(torrent.PiecePriorityNone)
				}
			}
		fmt.Printf("***\n")
}






///////////////////////////
func IsMainTorrent(magnet string) bool{
	return MainTorrent==magnet

}

/////////////////////////
/*
var SavedItems []string
func IsSavedItemWithMagnet(itempath string) bool{
	return  false
}
//func IsSavedItem(itempath string) bool{
func IsSavedItem(magnet string) bool{
	for _, tmpe:= range SavedItems {
		if tmpe==magnet {
    			return true
		}
	}
	return false
}
func AddSavedItem(itempath string){
	SavedItems=append(SavedItems,itempath)
}
func RemoveSavedItem(itempath string){
	SavedItems=removefromslice(SavedItems,itempath)
}
func removefromslice(slice []string, s string) []string {
	for i, tmpe:= range slice {
		if tmpe==s {
    	return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
*/
type SettingsType struct {
	LocalHostPort int
	SavedItems []SavedItemType
}

var Settings SettingsType

func LoadDefaultSettings(){
	Settings.LocalHostPort=666
	
}

func LoadSettings(){
   SettingsBytes, err := os.ReadFile("Settings.json") // just pass the file name
    if err != nil {
        fmt.Println("error:", err)
	LoadDefaultSettings()
	return
    }
	NewSettings:=new(SettingsType)
	uerr:=json.Unmarshal(SettingsBytes,NewSettings)
	if uerr != nil {
		fmt.Println("unmarshal error:", uerr)
		LoadDefaultSettings()
		return
	}
	Settings=*NewSettings

}

func SaveSettings(){
    f, err := os.Create("Settings.json")

    defer f.Close()

    //d2 := []byte{115, 111, 109, 101, 10}
	SettingsBytes, merr := json.Marshal(Settings)
	if merr != nil {
		fmt.Println("marshal error:", err)
		return
	}
    _, werr := f.Write(SettingsBytes)
       if werr != nil {
        fmt.Println("error:", werr)
	return
    }
    fmt.Printf("wrote settings\n")

}
/////////////////////////////////
type SavedItemType struct {
	Path string
 	Magnet string
}
//var SavedItems []SavedItemType
func IsSavedItemWithMagnet(magnet string) bool{
	for _, tmpe:= range Settings.SavedItems {
		if tmpe.Magnet==magnet {
    			return true
		}
	}
	return false
}
func IsSavedItem(itempath string) bool{
	for _, tmpe:= range Settings.SavedItems {
		if tmpe.Path==itempath {
    			return true
		}
	}
	return false
}
func AddSavedItem(itempath string,magnet string){
	var tmpsaveditem SavedItemType
	tmpsaveditem.Path=itempath
	tmpsaveditem.Magnet=magnet

	Settings.SavedItems=append(Settings.SavedItems,tmpsaveditem)
}
func RemoveSavedItem(itempath string){

	Settings.SavedItems=removefromslice(Settings.SavedItems,itempath)
}
func removefromslice(slice []SavedItemType, itempath string) []SavedItemType {
	for i, tmpe:= range slice {
		if tmpe.Path==itempath {
    			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}



