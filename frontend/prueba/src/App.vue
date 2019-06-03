<template>
  <div id="app">
    <div class="row">
      <div class="col">
        <b-container>
          <b-jumbotron  bg-variant="primary" text-variant="white" class="text-left" header="Search domain information" lead="">
            <!--<b-btn variant="primary" href="https://bootstrap-vue.js.org/">More Info</b-btn>-->
          </b-jumbotron>
          <b-form @reset="onReset">
            <b-form-group class="text-center" horizontal :label-cols="3" description="" label="Enter the domain">
              <b-form-input v-model.trim="dominioIn" v-on:click="onReset"></b-form-input>
            </b-form-group>
            <b-button variant="primary" v-on:click="fetch" v-bind:disabled="this.dominioIn === ''">Search Domain</b-button>
            <b-button variant="outline-secondary" v-on:click="searchList">List previous searches</b-button>
            <b-spinner variant="primary" class="float-right" label="Floated Right" v-show="showSpinner"></b-spinner>
          </b-form>
        </b-container>
      </div>
      <div class="col">
        <b-container>
          <b-card no-body>
            <b-nav pills slot="header" v-b-scrollspy:nav-scroller>
                <h3><b>Results</b></h3>
            </b-nav>
            <b-card-body
              id="nav-scroller"
              ref="content"
              style="position:relative; height:400px; overflow-y:scroll;"
            >
            <b-card-group tag="div">
              <div class="" v-for="host of hosts">
                <b-row>
                  <b-card style="">
                    <b-card-title class="text-left">
                    <b>Domain:</b> {{host.dominio}}
                    </b-card-title>
                      <b-card-text class="text-left">
                        <p>
                          <b>Title:</b> {{host.Title}}&nbsp;&nbsp;&nbsp;&nbsp;
                          <b>SSL grade:</b> {{host.SSL_grade}}&nbsp;&nbsp;&nbsp;&nbsp;
                          <b>Previous ssl grade:</b> {{host.Previous_ssl_grade}}&nbsp;&nbsp;&nbsp;&nbsp;
                          <b>Servers changed:</b> {{host.Servers_changed}}&nbsp;&nbsp;&nbsp;&nbsp;<br>
                          <b>Is down:</b> {{host.Is_down}}&nbsp;&nbsp;&nbsp;&nbsp;
                          <b>Logo:</b> <a v-bind:href="host.Logo" target="_blank">{{host.Logo}}</a>&nbsp;&nbsp;&nbsp;&nbsp;
                        </p>
                        <h4><b>Servers:</b></h4>
                      </b-card-text>
                      <b-card-text class="text-left">
                        <b-row>
                          <div class="" v-for="server of host.Servers">
                            <b-col>
                                  <b>IP Address:</b> {{server.Address}}</br>
                                  <b>SSL grade:</b> {{server.SSL_grade}}</br>
                                  <b>Country:</b> {{server.Country}}</br>
                                  <b>Owner:</b> {{server.Owner}}</br>
                            </b-col>
                          </div>
                        </b-row>
                      </b-card-text>
                  </b-card>
                </b-row>
              </br>
              </div>
            </b-card-group>
            </b-card-body>
          </b-card>
        </b-container>
      </div>
    </div>
    <br>
    <b-alert
      :show="dismissCountDown"
      dismissible
      variant="warning"
      @dismissed="dismissCountDown=0"
      @dismiss-count-down="countDownChanged"
    >
      <b>{{ error }}</b>
    </b-alert>
  </div>
</template>

<script>
import SearchMain from './components/SearchMain';
import axios from "axios";

export default {
  name: 'App',
  data: function() {
    return {
      dominioIn: '',
      queries: [],
      hosts : [],
      error: '',
      dismissSecs: 5,
      dismissCountDown: 0,
      showDismissibleAlert: false,
      show: true,
      spinner: false
    }
  },
  components: {
    SearchMain
  },
  methods:{
    getAndOrderData(res){
      this.hosts = [];
      var i;
      for(i = 0; i < res.data.Host.length; i++){
        var hostIn = new Object();
        hostIn.dominio = res.data.Host[i].Host;
        hostIn.SSL_grade = res.data.Host[i].SSL_grade;
        hostIn.Previous_ssl_grade = res.data.Host[i].Previous_ssl_grade;
        hostIn.Servers_changed = res.data.Host[i].Servers_changed;
        hostIn.Title = res.data.Host[i].Title;
        hostIn.Logo = res.data.Host[i].Logo;
        hostIn.Is_down = res.data.Host[i].Is_down;
        hostIn.Servers = res.data.Host[i].Servers;
        this.hosts.push(hostIn);
      }
    },
    fetch(evt){
      this.spinner = true;
      let result = axios.get("http://localhost:3000/v1/test/getInfoDomain/"+this.dominioIn)
        .then(res => {
          status = res.data.Status;
          if (status == "") {
            this.getAndOrderData(res);
            this.addQuery(this.dominioIn);
            this.spinner = false;
          } else {
            this.spinner = false;
            this.error=res.data.StatusMessage;
            console.log(this.error);
            this.dismissCountDown = this.dismissSecs;
          }
        })
        .catch(err => {
          console.log(err);
        })
    },
    searchList(){
      if(this.queries.length == 0){
        this.error="You have not done recent searches.";
        this.dismissCountDown = this.dismissSecs;
      } else {
        this.spinner = true;
        let result = axios.get("http://localhost:3000/v1/test/getInfoMultipleDomains/"+this.makeQuery())
          .then(res => {
            this.getAndOrderData(res);
            this.spinner = false;
          })
          .catch(err => {
            console.log(err);
          })
      }
    },
    addQuery(querieIn){
      if (this.queries.includes(querieIn.toLowerCase()) == false) {
        this.queries.push(querieIn.toLowerCase());
      }
    },
    makeQuery(){
      var i;
      var output = '';
      for(i = 0; i < this.queries.length; i++){
        output += this.queries[i]+" ";
      }
      return output;
    },
    countDownChanged(dismissCountDown) {
        this.dismissCountDown = dismissCountDown
      },
      onReset(evt){
        evt.preventDefault()
        this.dominioIn = ''
      }
  },
  computed:{
    showSpinner(){
      return this.spinner == true ? true : false
    }
  }
}
</script>

<style>
#app {
  font-family: 'Avenir', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  margin-top: 50px;
}
</style>
