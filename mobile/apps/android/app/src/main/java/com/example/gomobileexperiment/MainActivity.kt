package com.example.gomobileexperiment

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.unit.dp
import core.Core
import core.Greeter
import core.Profile
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            MaterialTheme {
                Surface(Modifier.fillMaxSize()) {
                    GreetingScreen()
                }
            }
        }
    }
}

sealed interface ProfileState {
    data object Idle : ProfileState
    data object Loading : ProfileState
    data class Loaded(val profile: Profile) : ProfileState
    data class Failed(val message: String) : ProfileState
}

@Composable
fun GreetingScreen() {
    var name by remember { mutableStateOf("Stevie") }
    val greeter: Greeter = remember { Core.newGreeter("Howzatt") }

    val freeFn = Core.hello(name)
    val viaStruct = greeter.greet(name)

    var profile by remember { mutableStateOf<ProfileState>(ProfileState.Idle) }
    val scope = rememberCoroutineScope()

    Column(
        modifier = Modifier.fillMaxSize().padding(24.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp),
    ) {
        Text("Go-powered greeting", style = MaterialTheme.typography.headlineMedium)
        OutlinedTextField(
            value = name,
            onValueChange = { name = it },
            label = { Text("Your name") },
            modifier = Modifier.testTag("nameField"),
        )
        Text("free function: $freeFn", style = MaterialTheme.typography.bodyLarge)
        Text("via struct:    $viaStruct", style = MaterialTheme.typography.bodyLarge)

        Button(
            onClick = {
                scope.launch {
                    profile = ProfileState.Loading
                    profile = try {
                        val p = withContext(Dispatchers.IO) { Core.fetchProfile("user-1") }
                        ProfileState.Loaded(p)
                    } catch (e: Exception) {
                        ProfileState.Failed(e.message ?: "unknown error")
                    }
                }
            },
            modifier = Modifier.testTag("loadProfileButton"),
        ) {
            Text("Load Profile")
        }

        val statusText = when (val s = profile) {
            ProfileState.Idle -> "(no profile loaded)"
            ProfileState.Loading -> "Loading…"
            is ProfileState.Loaded -> "Loaded: ${s.profile.name} (${s.profile.id})"
            is ProfileState.Failed -> "Error: ${s.message}"
        }
        Text(statusText, modifier = Modifier.testTag("profileStatus"))
    }
}
