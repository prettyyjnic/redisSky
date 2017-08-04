module.exports = {
  /**
   * Application configuration section
   * http://pm2.keymetrics.io/docs/usage/application-declaration/
   */
  apps : [

    // First application
    {
      name      : 'backend',
      script    : '/home/yjnic/Downloads/go/bin/go',
      cwd : '/mnt/hgfs/code/golang/src/github.com/prettyyjnic/redisSky/backend/bin',
      args: 'run start.go',
      watch: true,
      env: {
        "NODE_ENV": "production",
        "GOPATH": "/mnt/hgfs/code/golang",
        "GOROOT": "/home/yjnic/Downloads/go"
      }
    }
  ]
};
