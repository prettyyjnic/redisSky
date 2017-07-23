const routers = [{
    path: '/',
    meta: {
        title: 'redisSky manager'
    },
    component: (resolve) => require(['./views/index.vue'], resolve),    
    children: [
        {
            path: '/servers/edit/:serverid?',
            meta: {
                title: 'servers -redisSky manager'
            },
            component: (resolve) => require(['./views/server/edit.vue'], resolve)
        },
        {
            path: '/servers/list',
            meta: {
                title: 'servers list -redisSky manager'
            },
            component: (resolve) => require(['./views/server/list.vue'], resolve)
        },
        {
            path: '/system/configs',
            meta: {
                title: 'system - redisSky manager'
            },
            component: (resolve) => require(['./views/system.vue'], resolve)
        },
    	{
    		path: '/serverid/:serverid/keys',
    		meta: {
		        title: 'keys - redisSky manager'
		    },
		    component: (resolve) => require(['./views/keys.vue'], resolve),
		    children: [
		    	{
		    		path: '/serverid/:serverid/keys/:key',
		    		component: (resolve) => require(['./views/item.vue'], resolve)
		    	}
		    ]
    	}
    ]
}];
export default routers;