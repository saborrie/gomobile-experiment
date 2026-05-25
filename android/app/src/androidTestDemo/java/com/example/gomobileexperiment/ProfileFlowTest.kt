package com.example.gomobileexperiment

import androidx.compose.ui.test.assertTextEquals
import androidx.compose.ui.test.junit4.createAndroidComposeRule
import androidx.compose.ui.test.onAllNodesWithText
import androidx.compose.ui.test.onNodeWithTag
import androidx.compose.ui.test.performClick
import androidx.test.ext.junit.runners.AndroidJUnit4
import core.Core
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith

// Lives under androidTestDemo so it only compiles + runs against the demo
// flavor. Core.setScenario is defined only in the demo AAR; calling it from
// a prod-flavor build would fail at compile time, which is the design.
@RunWith(AndroidJUnit4::class)
class ProfileFlowTest {
    @get:Rule
    val rule = createAndroidComposeRule<MainActivity>()

    @Test
    fun happyPath_loadsProfile() {
        Core.setScenario("happy")

        rule.onNodeWithTag("loadProfileButton").performClick()

        rule.waitUntil(timeoutMillis = 5_000) {
            rule.onAllNodesWithText("Loaded:", substring = true)
                .fetchSemanticsNodes()
                .isNotEmpty()
        }
        rule.onNodeWithTag("profileStatus")
            .assertTextEquals("Loaded: Demo User (user-1)")
    }

    @Test
    fun notFoundScenario_showsError() {
        Core.setScenario("not-found")

        rule.onNodeWithTag("loadProfileButton").performClick()

        rule.waitUntil(timeoutMillis = 5_000) {
            rule.onAllNodesWithText("Error:", substring = true)
                .fetchSemanticsNodes()
                .isNotEmpty()
        }
        rule.onNodeWithTag("profileStatus")
            .assertTextEquals("Error: profile not found")
    }
}
