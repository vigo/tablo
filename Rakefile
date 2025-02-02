# frozen_string_literal: true

require 'English'

desc 'run test'
task :test do
  system %{ go test -failfast -v -coverprofile=coverage.out ./... }
  $CHILD_STATUS&.exitstatus || 1
rescue Interrupt
  0
end

desc 'show test coverage'
task :coverage do
  system %{ go test -v -coverprofile=coverage.out ./... && go tool cover -html=coverage.out }
  $CHILD_STATUS&.exitstatus || 1
rescue Interrupt
  0
end
