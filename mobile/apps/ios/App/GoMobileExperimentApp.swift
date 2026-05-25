import SwiftUI
import Core

@main
struct GoMobileExperimentApp: App {
    init() {
        // Point the Go core at the gRPC backend baked in at build time.
        // BackendEndpoint is set in Info.plist by xcodegen via the
        // INFOPLIST_KEY_BackendEndpoint setting in project.yml.
        let endpoint = (Bundle.main.object(forInfoDictionaryKey: "BackendEndpoint") as? String)
            ?? "localhost:7777"
        CoreSetEndpoint(endpoint)
    }

    var body: some Scene {
        WindowGroup {
            ContentView()
        }
    }
}
