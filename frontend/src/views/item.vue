<style scoped>
.offset-top-10{
    margin-top: 10px;
}
.offset-left-10{
    margin-left: 10px;
}
</style>
<template>
    <Row :gutter="16">
        <Col span="6">
            <span :style="{width: '20%'}">{{type}}: </span><Input v-model="key" :style="{width: '80%'}"></Input>
        </Col>
        <Col span="8">            
            <span>size: {{size}}</span>
            <span>TTL: <Input v-model="ttl" :style="{width: '60%'}"></Input></span>
        </Col>
        <Col span="10">
            <span :style="{float:'right'}">
                <Button>Reload</Button>
                <Button type="info">Set TTL</Button>
                <Button type="primary">Rename</Button>
                <Button type="warning">Delete</Button>
            </span>
        </Col>
    
        <Col span="24">
            <Row class="offset-top-10">
                <Col span="17">
                    <Table border height="200" size="small" :columns="columns" :data="data" :highlight-row="true" @on-row-click="viewData" @on-current-change="selectRow"></Table>
                </Col>
                <Col span="7">
                    <Button class="offset-top-10 offset-left-10" type="ghost" long><Icon type="plus-round" :color="'green'"></Icon>Add Row</Button>
                    <Button class="offset-top-10 offset-left-10" type="ghost" long @click="removeRow" :loading="removeBtnLoading">
                        <span v-if="!removeBtnLoading">
                            <Icon type="android-delete" :color="'red'"></Icon>Delete Row
                        </span>
                        <span v-else>Loading...</span>
                    </Button>
                    <Input class="offset-top-10 offset-left-10" v-model="searchKey" placeholder="please input field..."></Input>


                    <Col class="offset-top-10 offset-left-10" span="24">
                        <span>scan nums: </span>
                        <Input-number :max="size" :min="1" :step="1000" v-model="scanNums"></Input-number>
                    </Col>
                </Col>
            </Row>
        </Col>

        <Col span="24">
            <Row class="offset-top-10">
                <Col span="24">Score:</Col>
                <Col span="24"><Input v-model="score"></Input></Col>
            </Row>
        </Col>

        <Col span="24">
            <Row class="offset-top-10">
                <Col span="16">
                    <p>field({{bytes}} bytes): </p>
                </Col>
                <Col span="8">
                    <p>
                        <span>View as：</span>
                        <Select v-model="fieldType" :style="{width:'60%', float:'right'}">
                            <Option value="Text">Plain Text</Option>
                            <Option value="Json">Json</Option>
                        </Select>
                    </p>
                </Col>
            </Row>
            <Row class="offset-top-10">
                <Input v-model="field" type="textarea" :autosize="{minRows: 5,maxRows: 5}"></Input>
            </Row>
        </Col>

        <Col span="24">
            <Row class="offset-top-10">
                <Col span="16">
                    <p>Value({{bytes}} bytes): </p>
                </Col>
                <Col span="8">
                    <p>
                        <span>View as：</span>
                        <Select v-model="valueType" :style="{width:'60%', float:'right'}">
                            <Option value="Text">Plain Text</Option>
                            <Option value="Json">Json</Option>
                        </Select>
                    </p>
                </Col>
            </Row>
            <Row class="offset-top-10">
                <Input v-model="val" type="textarea" :autosize="{minRows: 10,maxRows: 50}"></Input>
            </Row>
        </Col>

        <Col span="24">
            <Row class="offset-top-10">
                <Button type="info" :style="{float:'right'}">Save</Button>
            </Row>
        </Col>
    </Row>
</template>

<script>
    export default {
        data(){
            return {
                key: "",
                type: "hash",
                field: "field",
                valueType: "Text",
                fieldType: "Text",
                searchKey: "",
                val: '{"hello":"world"}',
                size: 10000,
                bytes: 1,                
                score: 1,
                selectedRow: false,
                scanNums: 1000,
                removeBtnLoading: false,
                columns: [
                    {
                        title: 'row',
                        // key: 'row',
                        ellipsis: true,
                        type: 'index',
                        width: 60
                    },
                    {
                        title: 'field',
                        key: 'field',
                        ellipsis: true,
                    },
                    {
                        title: 'val',
                        key: 'val',
                        ellipsis: true,
                    },
                    {
                        title: 'score',
                        key: 'score',
                        sortable: true,
                        width: 90
                    }
                ],
                data: [{ field:"xxx", val:'{"hello":"val"}', score: 1},{ field:"xxx2", val:"val2", score: 111}],
                totalData: [],
                ttl: 1
            }
        },
        created () {
            // 组件创建完后获取数据，
            // 此时 data 已经被 observed 了
            this.fetchData();
        },
        watch: {
            // 如果路由有变化，会再次执行该方法
            '$route': 'fetchData',
            valueType: function(){
                if (this.valueType == 'Json') {
                    this.val = this.format(this.val)
                }else if(this.valueType == 'Text'){
                    this.val = this.format(this.val, true)
                }
            },
            fieldType: function(){
                if (this.fieldType == 'Json') {
                    this.field = this.format(this.field)
                }else if(this.fieldType == 'Text'){
                    this.field = this.format(this.field, true)
                }  
            },
            searchKey: function(){
                if (this.searchKey == "") {
                    this.data = this.totalData;
                }                
                var tmp = [];
                var reg = new RegExp(this.searchKey)
                for (var i = 0; i < this.totalData.length; i++) {
                    if (reg.test(this.totalData[i].field)) {
                        tmp.push(this.totalData[i]);
                    }
                }
                this.data = tmp;
            }
        },
        methods: {
            viewData: function (info){
                this.val = this.valueType == 'Json' ? this.format(info.val) : this.format(info.val, true);
                this.field = info.field;
            },
            fetchData(){
                this.key = this.$route.params.key;
                // ajax 获取数据
                this.totalData = this.data;
            },
            removeRow () {
                if (this.removeBtnLoading == true) {return;}                
                this.removeBtnLoading = true;
                // ajax 请求删除
                for (var i = 0; i < this.data.length ; i++) {
                    if (this.data[i]['field'] == this.selectedRow['field']) {
                        this.data.splice(i, 1);                        
                    }
                    if (this.totalData[i]['field'] == this.selectedRow['field']) {
                        this.totalData.splice(i, 1);                        
                    }
                }
                this.selectedRow = false;
                this.score = 0;
                this.field = "";
                this.val = "";
                this.removeBtnLoading = false;
            },
            selectRow (currentRow, oldCurrentRow){
                this.selectedRow = currentRow;
            },
            format: function format(txt,compress/*是否为压缩模式*/){/* 格式化JSON源码(对象转换为JSON文本) */
                // 参考 http://blog.csdn.net/macwhirr123/article/details/50686841
                var indentChar = '    ';     
                if(/^\s*$/.test(txt)){
                    return '';     
                }     
                try{var data=eval('('+txt+')');}     
                catch(e){     
                    return txt;     
                };     
                var draw=[],last=false,This=this,line=compress?'':'\n',nodeCount=0,maxDepth=0;     
                     
                var notify=function(name,value,isLast,indent/*缩进*/,formObj){     
                    nodeCount++;/*节点计数*/    
                    for (var i=0,tab='';i<indent;i++ )tab+=indentChar;/* 缩进HTML */    
                    tab=compress?'':tab;/*压缩模式忽略缩进*/    
                    maxDepth=++indent;/*缩进递增并记录*/    
                    if(value&&value.constructor==Array){/*处理数组*/    
                        draw.push(tab+(formObj?('"'+name+'":'):'')+'['+line);/*缩进'[' 然后换行*/    
                        for (var i=0;i<value.length;i++)     
                            notify(i,value[i],i==value.length-1,indent,false);     
                        draw.push(tab+']'+(isLast?line:(','+line)));/*缩进']'换行,若非尾元素则添加逗号*/    
                    }else if(value&&typeof value=='object'){/*处理对象*/    
                        draw.push(tab+(formObj?('"'+name+'":'):'')+'{'+line);/*缩进'{' 然后换行*/    
                        var len=0,i=0;     
                        for(var key in value)len++;     
                        for(var key in value)notify(key,value[key],++i==len,indent,true);     
                        draw.push(tab+'}'+(isLast?line:(','+line)));/*缩进'}'换行,若非尾元素则添加逗号*/    
                    }else{     
                        if(typeof value=='string')value='"'+value+'"';     
                        draw.push(tab+(formObj?('"'+name+'":'):'')+value+(isLast?'':',')+line);     
                    };     
                };     
                var isLast=true,indent=0;     
                notify('',data,isLast,indent,false);     
                return draw.join('');     
            }

        }

    }
</script>