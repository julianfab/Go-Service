package services;

import (
  "net/http"
  "bytes"
  "time"
  "log"
  "io/ioutil"
  "strings"
  "strconv"
  //"encoding/json"
  "github.com/go-chi/chi"
  "github.com/go-chi/render"
  "github.com/PuerkitoBio/goquery"
  "github.com/buger/jsonparser"
  _ "github.com/lib/pq"
  "database/sql"
);

type PostgresConn struct {
	Conn *sql.DB
}

type Response struct {
  Status string
  StatusMessage string
  Host []HostInfo
}

type HostInfo struct {
  Host string
  SSL_grade string
  Servers_changed bool
  Previous_ssl_grade string
  Logo string
  Title string
  Is_down bool
  Servers []ServerInfo
}

type ServerInfo struct {
  Address string
  SSL_grade string
  Country string
  Owner string
}

var myClient = &http.Client{Timeout: 10 * time.Second};
var postgresConn *sql.DB;
const urlSSLab string = "https://api.ssllabs.com/api/v3/analyze?host=";
const urlWhois string = "https://whois.arin.net/rest/";


func Routes(dbIn *sql.DB) *chi.Mux {
  postgresConn = dbIn;
  router := chi.NewRouter();
  router.Get("/getInfoDomain/{domain}", getInfoByDomain);
  router.Get("/getInfoMultipleDomains/{domains}", getInfoMultipleDomains);
  return router;
};

func getInfoMultipleDomains(w http.ResponseWriter, r *http.Request){
  var response Response;
  domainsIn := strings.ToLower(chi.URLParam(r,"domains"));
  domains := strings.Fields(domainsIn);
  for i:=0; i < len(domains); i++ {
    getInfoDomainByDB(&response, domains[i]);
  }
  render.JSON(w,r,response);
}

func getInfoByDomain(w http.ResponseWriter, r *http.Request){
  var response Response;
  response.Host = append(response.Host, HostInfo{});
  domain := strings.ToLower(chi.URLParam(r,"domain"));
  completeDomain(&response.Host[0], domain);
  getTitleAndLogo(&response);
  if response.Status == ""{
    getInfoSSLab(&response, domain);
    if response.Status == "" {
      if domainExist(response.Host[0].Host){
        response.Host[0].Servers_changed = serves_changed(response.Host[0]);
        updateInfo(&response.Host[0]);
      } else {
        response.Host[0].Previous_ssl_grade = response.Host[0].SSL_grade;
        makeRegister(response.Host[0]);
      }
    }
  }
  render.JSON(w,r,response)
}

func completeDomain(hostIn *HostInfo, domainIn string){
  if bytes.Index([]byte(domainIn), []byte("https://www.")) == -1 {
    hostIn.Host = "https://www."+domainIn;
  } else {
    if bytes.Index([]byte(domainIn), []byte("https://")) == -1 {
      hostIn.Host = "https://"+domainIn;
    }
  }
}

func getTitleAndLogo(responseIn *Response){
  res, err := http.Get(responseIn.Host[0].Host);
  if err != nil {
    responseIn.Status = "Error";
    responseIn.StatusMessage = "This domain can not be accessed";
  } else {
    defer res.Body.Close();
    if res.StatusCode != 200 {
      responseIn.Host[0].Is_down = true;
      responseIn.Status = "Error";
      responseIn.StatusMessage = "The searched domain is down";
    } else {
      doc, err := goquery.NewDocumentFromReader(res.Body);
      if err != nil {
        log.Fatal(err);
      }
      responseIn.Host[0].Title = doc.Find("head title").Text();
      var logo string;
      link := false;
      doc.Find("head link").Each(func(i int, s *goquery.Selection) {
        x, _ := s.Attr("rel");
        if bytes.Index([]byte(x), []byte("icon")) != -1 {
          link = true;
          if link {
            logoIn, _ := s.Attr("href");
            logo = logoIn;
          }
        }
      })
      if link != true {
        doc.Find("head meta").Each(func(i int, s *goquery.Selection) {
          x, _ := s.Attr("content");
          if bytes.Index([]byte(x), []byte("logo")) != -1 || bytes.Index([]byte(x), []byte("imag")) != -1 {
            if link != true {
              link = true;
              logoIn, _ := s.Attr("content");
              logo = logoIn;
            }
          }
        })
    }
      if bytes.Index([]byte(logo), []byte("https://")) == -1 {
        responseIn.Host[0].Logo = responseIn.Host[0].Host+logo;
      } else {
        responseIn.Host[0].Logo = logo;
      }
    }
  }
}

func getInfoDomainByDB(responseIn *Response, domainIn string){
  query, err := postgresConn.Query("SELECT id, name, servers_changed, ssl_grade, previous_ssl_grade, logo, title, is_down FROM dominio WHERE name = '"+domainIn+"';");
  if err != nil{
    log.Fatal("getInfoDomaiByDB",err);
  }
  defer query.Close();
  for query.Next(){
    var id, name, ssl_grade, previous_ssl_grade, logo, title string;
    var servers_changed, is_down bool;
    query.Scan(&id, &name, &servers_changed, &ssl_grade, &previous_ssl_grade, &logo, &title, &is_down);
    hostIn := HostInfo{Host:name, SSL_grade:ssl_grade, Servers_changed:servers_changed, Previous_ssl_grade:previous_ssl_grade, Logo:logo, Title:title, Is_down:is_down};
    getServersByDomainDB(&hostIn, id);
    responseIn.Host = append(responseIn.Host, hostIn);
  }
}

func getServersByDomainDB(hostIn *HostInfo, idDomainIn string){
  query, err := postgresConn.Query("SELECT address, ssl_grade, country, owner FROM server WHERE id_dominio = "+idDomainIn+";");
  if err != nil{
    log.Fatal("getServersByDomainDB ",err);
  }
  defer query.Close();
  for query.Next(){
    var address, ssl_grade, country, owner = "", "", "", "";
    query.Scan(&address, &ssl_grade, &country, &owner);
    serverIn := ServerInfo{Address:address, SSL_grade:ssl_grade, Country:country, Owner:owner};
    hostIn.Servers = append(hostIn.Servers, serverIn);
  }
}

func getInfoSSLab(responseIn *Response, domainIn string){
  follow := true;
  urlRequestSSLab := urlSSLab + domainIn;
  for follow {
    responseSSLab, err := myClient.Get(urlRequestSSLab);
    if err != nil {
      responseIn.Status = "Error";
      responseIn.StatusMessage = "The SSLAB service does not respond. Try again later.";
      break;
    }
    dataSSLab, _ := ioutil.ReadAll(responseSSLab.Body);
    defer responseSSLab.Body.Close();
    status := getJsonField(dataSSLab, "status");
    if strings.Compare(status, "ERROR") == 0 {
      responseIn.Status = "Error";
      responseIn.StatusMessage = "Unable to resolve domain name";
      break;
    }
    if strings.Compare(status, "READY") == 0 {
      responseIn.Host[0].Host = domainIn;
      responseIn.Host[0].Is_down = false;
      jsonparser.ArrayEach(dataSSLab, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
        ipAddressIn := getJsonField(value, "ipAddress");
        gradeIn := getJsonField(value, "grade");
        countryIn, ownerIn := getCountryAndOwnerByIp(ipAddressIn);
        serverIn := ServerInfo{Address:ipAddressIn, SSL_grade:gradeIn, Country:countryIn, Owner:ownerIn};
        responseIn.Host[0].Servers = append(responseIn.Host[0].Servers, serverIn);
      }, "endpoints");
      responseIn.Host[0].SSL_grade = getSSLGrade(responseIn.Host[0].Servers);
      break;
    } else {
        responseIn.Status = "Error";
        responseIn.StatusMessage = "The servers are being analyzed. Try again later.";
        break;
    }
  }
}

func getCountryAndOwnerByIp(ipAddressIn string) (country, owner string) {
  urlRequestOwner := urlWhois + "ip/" + ipAddressIn + ".json";
  responseOwner, err := myClient.Get(urlRequestOwner);
  if err != nil {
    owner, country = "undefined","undefined";
    return;
  }
  dataOwner, _ := ioutil.ReadAll(responseOwner.Body);
  defer responseOwner.Body.Close();
  owner = getJsonField(dataOwner, "net","orgRef","@name");
  orgRef := getJsonField(dataOwner, "net","orgRef","@handle");

  urlRequestCountry := urlWhois + "org/" + orgRef + ".json";
  responseCountry, err0 := myClient.Get(urlRequestCountry);
  if err0 != nil {
    country = "undefined";
    return;
  }
  dataCountry, _ := ioutil.ReadAll(responseCountry.Body);
  defer responseCountry.Body.Close();
  country = getJsonField(dataCountry, "org", "iso3166-1","code2","$");
  return;
}

func domainExist(domainIn string) bool {
  query, err := postgresConn.Query("SELECT name FROM dominio WHERE name = '"+domainIn+"';");
  if err != nil{
    log.Fatal("domainEx",err);
  }
  defer query.Close();
  var result string;
  for query.Next(){
    query.Scan(&result);
  }
  if result == domainIn{
    return true;
  } else {
    return false;
  }
}

func serves_changed(hostIn HostInfo) bool {
  var change bool;
  rows, err := postgresConn.Query("SELECT server.address FROM dominio, server WHERE  dominio.name='"+hostIn.Host+"';");
  if err != nil{
    log.Fatal("serves_changed",err);
  }
  defer rows.Close();
  for rows.Next(){
    var addressIn string;
    rows.Scan(&addressIn);
    change = addressNotEquals(hostIn, addressIn);
    if change {
      break;
    }
  }
  return change;
}

func makeRegister(hostIn HostInfo){
  _, err := postgresConn.Exec("INSERT INTO dominio (name, servers_changed, ssl_grade, previous_ssl_grade, logo, title, is_down) VALUES ('"+hostIn.Host+"',"+strconv.FormatBool(hostIn.Servers_changed)+",'"+hostIn.SSL_grade+"','"+hostIn.Previous_ssl_grade+"','"+hostIn.Logo+"', '"+hostIn.Title+"',"+strconv.FormatBool(hostIn.Is_down)+");");
  if err != nil{
    log.Panicf("make: ",err);
  }
  query, _ := postgresConn.Query("SELECT id FROM dominio WHERE name='"+hostIn.Host+"';");
  defer query.Close();
  var idDominio string;
  for query.Next(){
    query.Scan(&idDominio);
  }
  for i:=0; i<len(hostIn.Servers); i++{
    _, err2 := postgresConn.Exec("INSERT INTO server (address, ssl_grade, country, owner, id_dominio) VALUES ('"+hostIn.Servers[i].Address+"','"+hostIn.Servers[i].SSL_grade+"','"+hostIn.Servers[i].Country+"','"+hostIn.Servers[i].Owner+"',"+idDominio+");");
    if err2 != nil{
      log.Panicf("updateInfo: ",err);
    }
  }
}

func updateInfo(hostIn *HostInfo){
  query, _ := postgresConn.Query("SELECT id, previous_ssl_grade FROM dominio WHERE name='"+hostIn.Host+"';");
  defer query.Close();
  var idDominio, psgOld string;
  for query.Next(){
    query.Scan(&idDominio, &psgOld);
  }
  for i:=0; i<len(hostIn.Servers); i++{
    _, err:= postgresConn.Exec("DELETE FROM server WHERE address='"+hostIn.Servers[i].Address+"';");
    if err != nil{
      log.Panicf("updateInfo1: ",err);
    }
    _, err2 := postgresConn.Exec("INSERT INTO server (address, ssl_grade, country, owner, id_dominio) VALUES ('"+hostIn.Servers[i].Address+"','"+hostIn.Servers[i].SSL_grade+"','"+hostIn.Servers[i].Country+"','"+hostIn.Servers[i].Owner+"',"+idDominio+");");
    if err2 != nil{
      log.Panicf("updateInf2: ",err2);
    }
  }
  _, err3 := postgresConn.Exec("UPDATE dominio SET servers_changed="+strconv.FormatBool(hostIn.Servers_changed)+", ssl_grade='"+hostIn.SSL_grade+"', previous_ssl_grade='"+psgOld+"', logo='"+hostIn.Logo+"', title='"+hostIn.Title+"', is_down="+strconv.FormatBool(hostIn.Is_down)+" WHERE id="+idDominio+";");
  if err3 != nil{
    log.Panicf("updateInfo3: ",err3);
  }
  hostIn.Previous_ssl_grade = psgOld;
}

func addressNotEquals(hostIn HostInfo, addressIn string) (exit bool){
  for i:=0; i<len(hostIn.Servers); i++{
    if strings.Compare(addressIn, hostIn.Servers[i].Address) == 0{
      exit = false;
      break;
    } else {
      exit = true;
    }
  }
  return
}

func gradeToNumber(gradeIn string) (output int) {
  if strings.Compare(gradeIn, "M") == 0 {output=0;}
  if strings.Compare(gradeIn, "T") == 0 {output=1;}
  if strings.Compare(gradeIn, "F") == 0 {output=2;}
  if strings.Compare(gradeIn, "E") == 0 {output=3;}
  if strings.Compare(gradeIn, "D") == 0 {output=4;}
  if strings.Compare(gradeIn, "C") == 0 {output=5;}
  if strings.Compare(gradeIn, "B") == 0 {output=6;}
  if strings.Compare(gradeIn, "A-") == 0 {output=7;}
  if strings.Compare(gradeIn, "A") == 0 {output=8;}
  if strings.Compare(gradeIn, "A+") == 0 {output=9;}
  return;
}

func numberToGrade(numberIn int) (output string) {
  if numberIn == 0 {output="M";}
  if numberIn == 1 {output="T";}
  if numberIn == 2 {output="F";}
  if numberIn == 3 {output="E";}
  if numberIn == 4 {output="D";}
  if numberIn == 5 {output="C";}
  if numberIn == 6 {output="B";}
  if numberIn == 7 {output="A-";}
  if numberIn == 8 {output="A";}
  if numberIn == 9 {output="A+";}
  return;
}

func compareGrade(grade1, grade2 string) (gradeOut string){
  num1 := (gradeToNumber(grade1));
  num2 := (gradeToNumber(grade2));
  if num1 <= num2 {
    return numberToGrade(num1);
  } else {
    return numberToGrade(num2);
  }
}

func getSSLGrade(serversIn []ServerInfo) (SSL_grade string){
  if len(serversIn) == 0 {
    SSL_grade = "undefined";
  }
  if len(serversIn) == 1 {
    SSL_grade = serversIn[0].SSL_grade;
  }
  SSL_grade = serversIn[0].SSL_grade;
  for i := 1; i < len(serversIn); i++ {
    SSL_grade = compareGrade(SSL_grade, serversIn[i].SSL_grade);
  }
  return;
}

func getJsonField(json []byte, field ...string) string { //jsonparser
  v, _, _, e := jsonparser.Get(json, field...);
	if e != nil {
		return "";
	}
	if bytes.IndexByte(v, '\\') == -1 {
		return string(v);
	}
  result, _ := jsonparser.ParseString(v);
	return result;
}
