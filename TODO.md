- Download engine outside of the container, then --skip-engine-download inside container.
	Currently, the engine is re-downloaded inside the container. But flutter version inside and outside container may not match. And `flutter build bundle` is performed outside.
- Add a "Starting docker container" and "Docker container finished" before and after docker cmd run.
- `hover build --help` gives odd message "--skip-flutter-build-bundle flutter build bundle   Skip the flutter build bundle step."
- When using --docker for run no message at the end about hotreload/hotrestart gets printed and every letter I type appears on the console
