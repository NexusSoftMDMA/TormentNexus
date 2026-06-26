module.exports = {
  apps: [
    {
      name: "chorus-web",
      script: "node_modules/next/dist/bin/next",
      args: "start -p 5050 -H 127.0.0.1",
      cwd: __dirname,
      instances: 1,
      autorestart: true,
      watch: false,
      env: {
        NODE_ENV: "production",
        CHORUS_DAEMON_URL: "http://127.0.0.1:7707",
      },
    },
    {
      name: "chorus-daemon",
      script: "dist/daemon/index.js",
      cwd: __dirname,
      instances: 1,
      autorestart: true,
      watch: false,
      env: {
        NODE_ENV: "production",
        CHORUS_DAEMON_PORT: "7707",
        CHORUS_DATA_DIR: "/home/ubuntu/.chorus",
      },
    },
  ],
};
