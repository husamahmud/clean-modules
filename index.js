#!/usr/bin/env node
const { spawn } = require('child_process')
const path = require('path')

const binaryPath = path.join(__dirname, 'bin', 'drop-modules')
const currentDir = process.cwd()

const child = spawn(binaryPath, [currentDir], {
  stdio: 'inherit',
  cwd: currentDir,
})

child.on('error', (err) => {
  console.error('Failed to start subprocess.', err)
  process.exit(1)
})

child.on('close', (code) => {
  process.exit(code)
})
