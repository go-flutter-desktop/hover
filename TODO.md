- Download engine outside of the container, then --skip-engine-download inside container.
	Currently, the engine is re-downloaded inside the container. But flutter version inside and outside container may not match. And `flutter build bundle` is performed outside.
- When using --docker for run no message at the end about hotreload/hotrestart gets printed and every letter I type appears on the console
