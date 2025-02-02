#!/usr/bin/env node
const { exec } = require('child_process')
const path = require('path')

const binaryPath = path.join(__dirname, 'bin', 'drop-modules')

console.log('Executing binary at:', binaryPath)

exec(binaryPath, { stdio: 'inherit' }, (err, stdout, stderr) => {
  if (err) {
    console.error('Error executing binary:', err)
    console.error(`stderr: ${stderr}`)
    process.exit(1)
  }
  if (stdout) console.log(stdout)
})
