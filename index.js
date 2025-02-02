#!/usr/bin/env node
const { exec } = require('child_process')
const path = require('path')

const binaryPath = path.join(__dirname, 'bin', 'drop-modules')

exec(binaryPath, (err, stdout, stderr) => {
  if (err) {
    console.error(`Error: ${stderr}`)
    process.exit(1)
  }
  console.log(stdout)
})
