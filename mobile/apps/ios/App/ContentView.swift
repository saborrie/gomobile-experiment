import SwiftUI
import Core

enum ProfileState {
    case idle
    case loading
    case loaded(CoreProfile)
    case failed(String)
}

struct ContentView: View {
    @State private var name: String = "Stevie"
    @State private var profileState: ProfileState = .idle

    // gomobile generates the factory as a free function `CoreNewGreeter`
    // returning an optional. NewGreeter on the Go side never returns nil.
    private let greeter: CoreGreeter = CoreNewGreeter("Howzatt")!

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text("Go-powered greeting")
                .font(.title)

            TextField("Your name", text: $name)
                .textFieldStyle(.roundedBorder)
                .accessibilityIdentifier("nameField")

            Text("free function: \(CoreHello(name))")
            Text("via struct:    \(greeter.greet(name))")

            Button("Load Profile") {
                Task { await loadProfile() }
            }
            .accessibilityIdentifier("loadProfileButton")

            Text(statusText)
                .accessibilityIdentifier("profileStatus")

            Spacer()
        }
        .padding()
    }

    private var statusText: String {
        switch profileState {
        case .idle:
            return "(no profile loaded)"
        case .loading:
            return "Loading…"
        case .loaded(let profile):
            return "Loaded: \(profile.name) (\(profile.id_))"
        case .failed(let msg):
            return "Error: \(msg)"
        }
    }

    @MainActor
    private func loadProfile() async {
        profileState = .loading
        let result = await Task.detached { () -> Result<CoreProfile, Error> in
            var nsError: NSError?
            let profile = CoreFetchProfile("user-1", &nsError)
            if let err = nsError {
                return .failure(err)
            } else if let profile {
                return .success(profile)
            } else {
                return .failure(NSError(domain: "Core", code: 0,
                    userInfo: [NSLocalizedDescriptionKey: "nil profile"]))
            }
        }.value

        switch result {
        case .success(let profile):
            profileState = .loaded(profile)
        case .failure(let err):
            profileState = .failed(err.localizedDescription)
        }
    }
}
