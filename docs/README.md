# Fabric Documentation

Welcome to the Fabric documentation! This directory contains detailed guides and technical documentation for various features and components of Fabric.

## 📚 Available Documentation

### Core Features

**[rest-api.md](./rest-api.md)**
Complete REST API reference and interactive Swagger documentation. Covers all HTTP endpoints for chat completions, pattern management, contexts, sessions, authentication, and integration examples.

**[contexts-and-sessions-tutorial.md](./contexts-and-sessions-tutorial.md)**
Tutorial for using contexts and sessions to manage conversation state and reusable prompt data. Covers CLI usage, REST API endpoints, and practical workflows.

**[YouTube-Processing.md](./YouTube-Processing.md)**
Comprehensive guide for processing YouTube videos and playlists with Fabric. Covers transcript extraction, comment processing, metadata retrieval, and advanced yt-dlp configurations.

**[Using-Speech-To-Text.md](./Using-Speech-To-Text.md)**
Documentation for Fabric's speech-to-text capabilities using OpenAI's Whisper models. Learn how to transcribe audio and video files and process them through Fabric patterns.

**[Automated-Changelog-Usage.md](./Automated-Changelog-Usage.md)**
Complete guide for developers on using the automated changelog system. Covers the workflow for generating PR changelog entries during development, including setup, validation, and CI/CD integration.

### User Interface & Experience

**[Desktop-Notifications.md](./Desktop-Notifications.md)**
Guide to setting up desktop notifications for Fabric commands. Useful for long-running tasks and multitasking scenarios with cross-platform notification support.

**[Shell-Completions.md](./Shell-Completions.md)**
Instructions for setting up intelligent tab completion for Fabric in Zsh, Bash, and Fish shells. Includes automated installation and manual setup options.

**[Gemini-TTS.md](./Gemini-TTS.md)**
Complete guide for using Google Gemini's text-to-speech features with Fabric. Covers voice selection, audio generation, and integration with Fabric patterns.

### Setup & Configuration

**[GitHub-Models-Setup.md](./GitHub-Models-Setup.md)**
Comprehensive setup guide for using GitHub Models with Fabric. Covers authentication, model selection, and integration with GitHub's AI model marketplace.

**[i18n.md](./i18n.md)**
Internationalization implementation guide. Covers locale management, translation workflows, and adding new language support to Fabric.

**[i18n-variants.md](./i18n-variants.md)**
BCP 47 locale normalization and language variant handling. Documents support for regional variations like Brazilian Portuguese (pt-BR) and European Portuguese (pt-PT).

### Development & Architecture

**[Automated-ChangeLog.md](./Automated-ChangeLog.md)**
Technical documentation outlining the automated CHANGELOG system architecture for CI/CD integration. Details the infrastructure and workflow for maintainers.

**[Project-Restructured.md](./Project-Restructured.md)**
Project restructuring plan and architectural decisions. Documents the transition to standard Go conventions and project organization improvements.

**[NOTES.md](./NOTES.md)**
Development notes on refactoring efforts, model management improvements, and architectural changes. Includes technical details on vendor and model abstraction.

**[Go-Updates-September-2025.md](./Go-Updates-September-2025.md)**
Migration notes and architectural changes from the Python to Go rewrite. Documents major updates, breaking changes, and new capabilities introduced in September 2025.

### Community & Policies

**[CONTRIBUTING.md](./CONTRIBUTING.md)**
Contribution guidelines, development setup, and best practices for contributing to Fabric. Essential reading for all contributors.

**[CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md)**
Community code of conduct outlining expected behavior and guidelines for participation in the Fabric project.

**[SECURITY.md](./SECURITY.md)**
Security policy and vulnerability reporting procedures. Learn how to responsibly disclose security issues.

**[SUPPORT.md](./SUPPORT.md)**
Support resources and help channels. Find answers to common questions and learn where to get help.

### Audio Resources

**[voices/README.md](./voices/README.md)**
Index of Gemini TTS voice samples demonstrating different AI voice characteristics available in Fabric.

## 🗂️ Additional Resources

### Configuration Files

- `./notification-config.yaml` - Example notification configuration

### Images

- `images/` - Screenshots and visual documentation assets
  - `fabric-logo-gif.gif` - Animated Fabric logo
  - `fabric-summarize.png` - Screenshot of summarization feature
  - `svelte-preview.png` - Web interface preview

## 🚀 Quick Start

New to Fabric? Start with these essential docs:

1. **[../README.md](../README.md)** - Main project README with installation and basic usage
2. **[rest-api.md](./rest-api.md)** - REST API reference with interactive Swagger UI
3. **[contexts-and-sessions-tutorial.md](./contexts-and-sessions-tutorial.md)** - Learn to manage conversation state
4. **[Shell-Completions.md](./Shell-Completions.md)** - Set up tab completion for better CLI experience
5. **[YouTube-Processing.md](./YouTube-Processing.md)** - Learn one of Fabric's most popular features

## 🔧 For Contributors

Contributing to Fabric? These docs are essential:

1. **[CONTRIBUTING.md](./CONTRIBUTING.md)** - Contribution guidelines and development setup
2. **[CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md)** - Community guidelines and expected behavior
3. **[Automated-Changelog-Usage.md](./Automated-Changelog-Usage.md)** - Required workflow for PR submissions
4. **[Project-Restructured.md](./Project-Restructured.md)** - Understanding project architecture
5. **[SECURITY.md](./SECURITY.md)** - Security policy and vulnerability reporting
6. **[NOTES.md](./NOTES.md)** - Current development priorities and patterns

## 📝 Documentation Standards

When adding new documentation:

- Use clear, descriptive filenames
- Include practical examples and use cases
- Update this README index with your new docs
- Follow the established markdown formatting conventions
- Test all code examples before publication
