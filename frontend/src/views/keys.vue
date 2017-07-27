<style scoped>
.offset-top-10{
    margin-top: 10px;
}
.overflow-hidden{
    text-overflow: ellipsis;    
    overflow: hidden;
}
.overflow-y-show{
    overflow-y: auto;
}
</style>
<template>
    <Row :gutter="16">
        <Col span="5" class="left">
            <Card>
                <div class="offset-top-10">
                    <Select v-model="serverdb" class="width100">
                        <Option v-for="item in server.dbNums" :value="(item - 1)" :key="(item-1)">
                        db{{ item -1}}
                        </Option>
                    </Select>
                </div>
                <div class="offset-top-10">
                    <Input v-model="inputKey" placeholder="redis key..." class="width100"></Input>
                </div>

                <div class="overflow-y-show" ref="keysCard" :style="{ 'height': keysCardHeight}" >                    
                    <ul>
                        <li class="overflow-hidden" v-for="item in keys"><span class="layout-text" ><router-link :to="getLink(item)">{{item}}</router-link></span></li>
                    </ul>
                </div>
            </Card>
        </Col>
        <Col span="19">
            <Card>
                <router-view></router-view>
            </Card>
        </Col>
    </Row>
</template>

<script>
    export default {
        data(){
            return {
                inputKey: "",
                serverdb: 0,
                keys: this.getKeys(),
            }
        },
        
        computed: {
            keysCardHeight(){
                return window.innerHeight - 260 +"px";
            },
            serverid : function(){
                return parseInt(this.$route.params.serverid);
            },
            server(){
                for (var i = this.$store.state.servers.length - 1; i >= 0; i--) {
                    if (this.$store.state.servers[i]["id"] == this.serverid) {
                        return this.$store.state.servers[i];
                    }
                }
                return {dbNums:0, id:0}
            }
        },
        watch: {
            '$route': 'reload',
            // 如果 question 发生改变，这个函数就会运行
            inputKey () {
                this.getKeys();
            },
            serverdb (){
                this.getKeys();                
            }
        },
        methods: {
            getKeys: _.debounce(function(){
                var info = {}
                info.serverid = this.server.id;
                info.db = parseInt( this.serverdb );
                info.data = this.inputKey;
                this.$socket.emit("ScanKeys", info)
            }, 200),
            reload:  function(newRouter, oldRouter){
                if (this.serverid != this.$route.params.serverid || this.serverdb != this.$route.params.db) {
                    this.inputKey = "";
                    this.serverid = parseInt( this.$route.params.serverid );
                    this.getKeys();
                }
            },
            getLink(item){
                return '/serverid/'+ this.serverid + '/db/' + (this.serverdb ? parseInt(this.serverdb) : 0) +'/key/'+ item;
            }
        },
        socket:{
            events:{
                LoadKeys(keys){
                    this.keys = keys
                },
                ReloadKeys(){
                    this.getKeys();
                }
            }
        }

       
    }
</script>