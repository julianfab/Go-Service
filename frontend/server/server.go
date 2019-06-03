package server;

import (
  "net/http"
  "bytes"
  "time"
  "log"
  "io/ioutil"
  "strings"
  //"html/template"

  //"encoding/json"
  "github.com/go-chi/chi"
  "github.com/go-chi/render"
  "github.com/buger/jsonparser"
);

type StatusMessage struct {
  Status string
}

type HostInfo struct {
  Host string
  Servers_changed bool
  Previous_ssl_grade string
  Logo string
  Tittle string
  Is_down bool
  Servers []ServerInfo
}

type ServerInfo struct {
  Host string
  Address string
  SSL_grade string
  Country string
  Owner string
}

type Test struct {
	ID 	string `json:"id,omitempty"`;
};

func Routes() *chi.Mux {
  router := chi.NewRouter();
  router.Get("/", getAllTest);
  router.Get("/{id}", getTestById);
  router.Get("/json/{domain}", getServersByDomain);
  //router.Get("/json/test",getJsonTest);
  router.Get("/json/test1",test);
  router.Get("/domain/{domain}",getInfo);
  return router;
};

var tests []Test;
var hostsInfo []HostInfo;
var status StatusMessage;


func getAllTest(w http.ResponseWriter, r *http.Request){
  tests = append(tests, Test{ID:"2", });
  tests = append(tests, Test{ID:"22"});
  tests = append(tests, Test{ID:"23"});
  render.JSON(w,r,tests);
}

func getTestById(w http.ResponseWriter, r *http.Request)  {
  testID := chi.URLParam(r,"id");
  test1 := Test{
    ID: testID,
  }
  render.JSON(w,r,test1);
};

func getServersByDomain(w http.ResponseWriter, r *http.Request) {
  log.Printf("Aquix3")
  domain := chi.URLParam(r,"domain");
  var buffer bytes.Buffer;
  buffer.WriteString("https://api.ssllabs.com/api/v3/analyze?host=");
  buffer.WriteString(domain);
  test3 := Test{
    ID: buffer.String(),
  }
  render.JSON(w,r,test3);
}

var myClient = &http.Client{Timeout: 10 * time.Second}
/*
func getJson(target interface{}) error {
    //r, err := myClient.Get("https://api.ssllabs.com/api/v3/analyze?host=google.com")
    r, err := myClient.Get("http://http://whois.arin.net/rest/ip/34.192.0.0.json")
    //https://whois.arin.net/rest/org/AT-88-Z.json
    if err != nil {
        return err
    }
    //var host Host
    byteValue, _ := ioutil.ReadAll(r.Body)
    json.Unmarshal(byteValue, &host)
    log.Printf(string(byteValue))
    log.Printf("Puerto")
    log.Printf(host.Port)
    defer r.Body.Close()
    return json.NewDecoder(r.Body).Decode(target)
}*/



func getCountryAndOwnerByIp(ipAddressIn string) (country, owner string) {
  log.Printf("Entro al Segundo")
  var urlGetOwner bytes.Buffer;
  urlGetOwner.WriteString("https://whois.arin.net/rest/ip/");
  urlGetOwner.WriteString(ipAddressIn);
  urlGetOwner.WriteString(".json");
  urlOwner := urlGetOwner.String();
  log.Printf(urlOwner)
  req, err := myClient.Get(urlOwner);
  log.Printf("Antes del error")
  if err != nil {
      log.Panicf(err.Error())
  }
  log.Printf("Paso el error")
  dataIp, _ := ioutil.ReadAll(req.Body);
  defer req.Body.Close();
  log.Printf(string(dataIp))
  owner = getJsonField(dataIp, "net", "orgRef","@name")
  log.Printf(owner)
  log.Printf("Pasoooooopls")
  key := getJsonField(dataIp, "net", "orgRef","@handle")
  log.Printf(key)

  var urlGetCountry bytes.Buffer;
  urlGetCountry.WriteString("https://whois.arin.net/rest/org/");
  urlGetCountry.WriteString(key);
  urlGetCountry.WriteString(".json");
  urlCountry := urlGetCountry.String();
  req, err1 := myClient.Get(urlCountry);
  log.Printf("Nuevo error")
  if err1 != nil {
      log.Panicf(err.Error())
  }
  log.Printf("Si paso 0:")

  dataOrg, err2 := ioutil.ReadAll(req.Body);
  if err2 != nil {
      log.Panicf(err.Error())
  }
  defer req.Body.Close();
  log.Printf("Es este?")
  country = getJsonField(dataOrg, "org", "iso3166-1","code2","$")
  log.Printf(country)
  log.Printf("Si es")

  return
}

func getJsonField(json []byte, field ...string) string {
  //Sacado de el repo jsonParser
  v, _, _, e := jsonparser.Get(json, field...)
	if e != nil {
		return ""
	}
	// If no escapes return raw conten
	if bytes.IndexByte(v, '\\') == -1 {
		return string(v)
	}
  result, _ := jsonparser.ParseString(v)
	return result
  /*result, _ := jsonparser.GetString(json, field);
  return result;*/
}

func getServersInfo(domain string) bool{
  log.Printf("Aquix2")
  var urlSSLab bytes.Buffer;
  urlSSLab.WriteString("https://api.ssllabs.com/api/v3/analyze?host=");
  urlSSLab.WriteString(domain);
  urlReq := urlSSLab.String()
  req, err := myClient.Get(urlReq);
  log.Printf("Antes del error, AQUIX2")
  if err != nil {
      log.Panicf(err.Error())
  }
  log.Printf("paso")
  dataObtained, _ := ioutil.ReadAll(req.Body);
  //defer req.Body.Close();
  statusM := getJsonField(dataObtained,"statusMessage")
  if statusM != "" {
    status.Status = "No se encuentra información con el Dominio buscado."
    return false
  } else {
    status := getJsonField(dataObtained, "status")
    log.Printf("EL STATUS : ", status)
    if status == "READY" {
      hostInfoIn := HostInfo{}
      hostInfoIn.Host = domain;
      jsonparser.ArrayEach(dataObtained, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
        ipAddressIn, err := jsonparser.GetString(value, "ipAddress")
        gradeIn, err := jsonparser.GetString(value, "grade")
        log.Printf("EL IPADRES", ipAddressIn)
        log.Printf("EL grades", gradeIn)
        log.Printf("ANTES DEL ERROR DEL IF")
        countryIn := ""
        ownerIn := ""
        countryIn, ownerIn = getCountryAndOwnerByIp(ipAddressIn);
        log.Printf("DEPUSES DEL ERROR DEL IF")
        serverIn := ServerInfo{Host:domain, Address:ipAddressIn, SSL_grade:gradeIn, Country:countryIn, Owner:ownerIn}
        hostInfoIn.Servers = append(hostInfoIn.Servers, serverIn);
      }, "endpoints")
      hostsInfo = append(hostsInfo, hostInfoIn)
      return true
    }
    return false
  }
}

func getInfo(w http.ResponseWriter, r *http.Request){
  domain := chi.URLParam(r,"domain");
  log.Printf("Aqui comenzo")
  if getServersInfo(domain) {
    render.JSON(w,r,hostsInfo);
  } else {
    render.JSON(w,r,status);
  }

}

func test(w http.ResponseWriter, r *http.Request) {


  domain := strings.ToLower("TRUORA.com")

  var urlSSLab bytes.Buffer;
  urlSSLab.WriteString("https://api.ssllabs.com/api/v3/analyze?host=");
  urlSSLab.WriteString(domain);
  urlReq := urlSSLab.String()
  req, err := myClient.Get(urlReq);
  /*prueba := "https://api.ssllabs.com/api/v3/analyze?host=google.com"
  req, err := myClient.Get(prueba)*/
  if err != nil {
      log.Panicf(err.Error())
  }
  byteValue, _ := ioutil.ReadAll(req.Body)
  str := getJsonField(byteValue,"host")
  log.Printf(str)
  status := getJsonField(byteValue, "status")
  log.Printf(status)
  hostInfoIn := HostInfo{}
  hostInfoIn.Host = domain;
    jsonparser.ArrayEach(byteValue, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
      ipAddressIn := getJsonField(value, "ipAddress")
      log.Printf(ipAddressIn)
      gradeIn := getJsonField(value, "grade")
      countryIn, ownerIn := getCountryAndOwnerByIp(ipAddressIn);
      serverIn := ServerInfo{Host:domain, Address:ipAddressIn, SSL_grade:gradeIn, Country:countryIn, Owner:ownerIn}
      hostInfoIn.Servers = append(hostInfoIn.Servers, serverIn);
      /*val, err := jsonparser.GetString(value, "ipAddress")
      valor := getJsonField(value, "ipAddress")
       log.Printf(valor)
       log.Printf(val)*/
    }, "endpoints")
  hostsInfo = append(hostsInfo, hostInfoIn)
  render.JSON(w,r,hostsInfo)

/*  switch r.Method {
  case "GET":
    template, err := template.ParseFiles("index.html")
    if err != nil{
      log.Printf("PAGINA NO ENCONTRADA")
    } else {
      template.Execute(w, nil)
    }

  }*/

  //render.JSON(w,r,string(byteValue))
}

/*func getJsonTest(w http.ResponseWriter, r *http.Request){
  var test Host;
  getJson(test);
  render.JSON(w,r,test);

}*/

/*func getServersByDomain1(domain string) (Test, error) {
}*/
