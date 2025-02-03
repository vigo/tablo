# frozen_string_literal: true

require 'English'

task :command_exists, [:command] do |_, args|
  abort "#{args.command} doesn't exists" if `command -v #{args.command} > /dev/null 2>&1 && echo $?`.chomp.empty?
end

task :is_repo_clean do
  next if ENV['BYPASS_REPO_CLEAN']

  abort 'please commit your changes first!' unless `git status -s | wc -l`.strip.to_i.zero?
end

task :has_bump_my_version do
  Rake::Task['command_exists'].invoke('bump-my-version')
end


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

AVAILABLE_REVISIONS = %w[major minor patch].freeze
desc "bump version, default: patch, available: #{AVAILABLE_REVISIONS.join(',')}"
task :bump, [:revision] => [:has_bump_my_version] do |_, args|
  args.with_defaults(revision: 'patch')

  unless AVAILABLE_REVISIONS.include?(args.revision)
    abort "Please provide valid revision: #{AVAILABLE_REVISIONS.join(',')}"
  end

  system %{ bump-my-version bump #{args.revision} }
  exit $CHILD_STATUS.exitstatus unless $CHILD_STATUS.exitstatus.zero?
rescue Interrupt
  0
end
