<style scoped>
    .layout{
        border: 1px solid #d7dde4;
        background: #f5f7f9;
        position: relative;
        border-radius: 4px;
        overflow: hidden;        
    }
    .layout-breadcrumb{
        padding: 10px 15px 0;
    }
    .layout-content{
        min-height: 200px;
        margin: 15px;
        overflow: hidden;
        background: #fff;
        border-radius: 4px;
    }
    .layout-content-main{
        padding: 10px;
    }
    .layout-copy{
        text-align: center;
        padding: 10px 0 20px;
        color: #9ea7b4;
    }
    .layout-menu-left{
        background: #464c5b;
    }
    .layout-header{
        height: 60px;
        background: #fff;
        box-shadow: 0 1px 1px rgba(0,0,0,.1);
    }
    .layout-logo-left{
        width: 90%;
        height: 30px;
        background: #5b6270;
        border-radius: 3px;
        margin: 15px auto;
    }
    .layout-ceiling-main a{
        color: #9ba7b5;
    }
    .layout-hide-text .layout-text{
        display: none;
    }
    .ivu-col{
        transition: width .2s ease-in-out;
    }
</style>
<template>
    <div class="layout" :class="{'layout-hide-text': spanLeft < 3}" :style="{ 'height': windowHeight}">
        <Row type="flex" :style="{ 'height': windowHeight}">
            <i-col :span="spanLeft" class="layout-menu-left">
                <Menu active-name="serverList" theme="dark" width="auto" :accordion="true">
                    <Menu-group title="System">
                        <Menu-item name="serverList">
                            <Icon type="social-reddit"></Icon>
                            <span class="layout-text" ><router-link to='/servers/list'>server list</router-link></span>
                        </Menu-item>
                        <Menu-item name="addServer">
                            <Icon type="plus-circled"></Icon>
                            <span class="layout-text" ><router-link to='/servers/edit'>add server</router-link></span>
                        </Menu-item>
                        <Menu-item name="settings">
                            <Icon type="settings"></Icon>
                            <span class="layout-text" ><router-link to='/system/configs'>configs</router-link></span>
                        </Menu-item>
                    </Menu-group>

                    <Menu-group title="Servers">
                        <Menu-item :name="'Servers'+item.id" v-for="item in servers">
                            <Icon type="ios-navigate"></Icon>
                            <Tooltip :content="item.host" placement="right-end">
                                <span class="layout-text" >
                                    <router-link :to="'/serverid/'+ item.id +'/keys'">{{item.name}}</router-link>
                                </span>
                            </Tooltip>
                        </Menu-item>
                    </Menu-group>

                </Menu>
            </i-col>
            <i-col :span="spanRight">
                <div class="layout-header">
                    <i-button type="text" @click="toggleClick">
                        <Icon type="navicon" size="32"></Icon>
                    </i-button>
                </div>
               
                <div class="layout-content">
                    <div class="layout-content-main">
                        <router-view></router-view>
                    </div>
                </div>
                <div class="layout-copy">
                    2017-2017 &copy; sky
                </div>
            </i-col>
        </Row>
    </div>
</template>
<script>
    export default {
        data () {
            return {
                spanLeft: 3,
                spanRight: 21,
                // servers: {}
            }
        },
        created () {
            // 组件创建完后获取数据，
            // 此时 data 已经被 observed 了
            this.$socket.emit("QueryServers")
        },
        computed: {
            iconSize () {
                return this.spanLeft === 3 ? 14 : 22;
            },
            windowHeight(){
                return window.innerHeight+"px";
            },
            servers(){
                return this.$store.state.servers;
            }
        },
        methods: {
            toggleClick () {
                if (this.spanLeft === 3) {
                    this.spanLeft = 1;
                    this.spanRight = 23;
                } else {
                    this.spanLeft = 3;
                    this.spanRight = 21;
                }
            }
        },
        socket:{
            events:{
                
            }
        }
    }
</script>