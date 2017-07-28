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
.hidden{
    display: none;
}
.width100{
    width: 100% !important;
}
</style>
<style lang="less">
.vertical-center-modal{
    display: flex;
    align-items: center;
    justify-content: center;

    .ivu-modal{
        top: 0;
    }
}
</style>
<template>
    <Row :gutter="16">
        <Col span="5" class="left">
            <Card>
                <Button type="success" long @click="addKeyMoal = true">Add Key</Button>
                <Modal
                    title="Add Key"
                    v-model="addKeyMoal"
                    @on-ok="addKey"
                    class-name="vertical-center-modal">
                    <Form :model="newItem" :label-width="80">
                        <Form-item label="key" >
                            <Input v-model="newItem.key" type="text" placeholder="please input key..."></Input>
                        </Form-item>
                        <Form-item label="type" >
                            <Select v-model="newItem.t" class="width100">
                                <Option value="string">String</Option>
                                <Option value="list">List</Option>
                                <Option value="set">Set</Option>
                                <Option value="zset">Zset</Option>
                                <Option value="hash">Hash</Option>
                            </Select>
                        </Form-item>
                        <Form-item
                            :class="{'hidden': newItem.t != 'set' && newItem.t != 'list'}"
                            v-for="(item, index) in newItem.listVal"
                            :label="'val' + (index + 1)"
                            >
                            <Row>
                                <Col span="18">
                                    <Input type="text" v-model="newItem.listVal[index]" placeholder="please input ..."></Input>
                                </Col>
                                <Col span="4" offset="1">
                                    <Button type="ghost" @click="handleRemoveList(index)">del</Button>
                                </Col>
                            </Row>
                        </Form-item>
                        <Form-item :class="{'hidden': newItem.t != 'set' && newItem.t != 'list'}">
                            <Row>
                                <Col span="12">
                                    <Button type="dashed" long @click="handleAddList" icon="plus-round">add</Button>
                                </Col>
                            </Row>
                        </Form-item>
                        <Form-item
                            :class="{'hidden': newItem.t != 'hash' }"
                            v-for="(item, index) in newItem.hashVal"
                            :key="index"
                            :label="'val' + (index + 1)"
                            >
                            <Row>
                                <Col span="9">
                                    <Input type="text" v-model="newItem.hashVal[index].field" placeholder="please input field..."></Input>
                                </Col>
                                <Col span="9" offset="1">
                                    <Input type="text" v-model="newItem.hashVal[index].val" placeholder="please input val..."></Input>
                                </Col>
                                <Col span="4" offset="1">
                                    <Button type="ghost" @click="handleRemoveHash(index)">del</Button>
                                </Col>
                            </Row>
                        </Form-item>
                        <Form-item :class="{'hidden': newItem.t != 'hash' }">
                            <Row>
                                <Col span="12">
                                    <Button type="dashed" long @click="handleAddHash" icon="plus-round">add</Button>
                                </Col>
                            </Row>
                        </Form-item>
                        <Form-item
                            :class="{'hidden': newItem.t != 'zset' }"
                            v-for="(item, index) in newItem.zsetVal"
                            :key="index"
                            :label="'val' + (index + 1)"
                            >
                            <Row>
                                <Col span="9">
                                    <Input type="text" v-model="newItem.zsetVal[index].val" placeholder="please input val..."></Input>
                                </Col>
                                <Col span="9" offset="1">
                                    <Input type="text" v-model="newItem.zsetVal[index].score" placeholder="please input score..."></Input>
                                </Col>
                                <Col span="4" offset="1">
                                    <Button type="ghost" @click="handleRemoveZset(index)">del</Button>
                                </Col>
                            </Row>
                        </Form-item>
                        <Form-item :class="{'hidden': newItem.t != 'zset' }">
                            <Row>
                                <Col span="12">
                                    <Button type="dashed" long @click="handleAddZset" icon="plus-round">add</Button>
                                </Col>
                            </Row>
                        </Form-item>
                        <Form-item label="value" :class="{'hidden': newItem.t != 'string' }">
                            <Input v-model="newItem.stringVal" type="textarea" :autosize="{minRows: 4,maxRows: 6}" placeholder="please input string val..."></Input>
                        </Form-item> 
                    </Form>
                </Modal>
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
                addKeyMoal: false,
                keys: this.getKeys(),
                newItem: {
                    t: 'string',
                    stringVal: '',
                    key: '',
                    listVal: [''],
                    zsetVal: [{val:'', score:0}],
                    hashVal: [{field:'', val:''}]
                }
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
            handleAddList () {
                this.newItem.listVal.push('');
            },
            handleRemoveList (index) {
                this.newItem.listVal.splice(index, 1);
            },
            handleAddHash () {
                this.newItem.hashVal.push({field:'', val:''});
            },
            handleRemoveHash (index) {
                this.newItem.hashVal.splice(index, 1);
            },
            handleAddZset () {
                this.newItem.zsetVal.push({val:'', score:0});
            },
            handleRemoveZset (index) {
                this.newItem.zsetVal.splice(index, 1);
            },
            reload:  function(newRouter, oldRouter){
                if (this.serverid != this.$route.params.serverid || this.serverdb != this.$route.params.db) {
                    this.inputKey = "";
                    this.serverid = parseInt( this.$route.params.serverid );
                    this.getKeys();
                }
            },
            getLink(item){
                return '/serverid/'+ this.serverid + '/db/' + (this.serverdb ? parseInt(this.serverdb) : 0) +'/key/'+ item;
            },
            addKey(){
                // add one key to redis
                var info = {};
                info.db = this.serverdb ? parseInt(this.serverdb) : 0;
                info.serverid = parseInt( this.server.id );
                info.data = {}
                info.data.key = this.newItem.key;
                if (this.newItem.t == 'string') {
                    info.data.val = this.newItem.stringVal;
                }else if(this.newItem.t == 'set' || this.newItem.t == 'list'){
                    info.data.val = this.newItem.listVal;
                }else if(this.newItem.t == 'zset'){
                    info.data.val = this.newItem.zsetVal;
                }else if(this.newItem.t == 'hash'){
                    info.data.val = this.newItem.hashVal;
                }else{
                    this.$Notice.error({
                        title: 'Error',
                        desc: "unknown type : " + this.newItem.t
                    });
                    return;
                }
                info.data.t = this.newItem.t;
                this.$socket.emit("AddKey", info);
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