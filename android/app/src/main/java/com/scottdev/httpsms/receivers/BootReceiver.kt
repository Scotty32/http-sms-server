package com.scottdev.httpsms.receivers

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import com.scottdev.httpsms.Constants
import com.scottdev.httpsms.Settings
import com.scottdev.httpsms.services.StickyNotificationService
import timber.log.Timber


class BootReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent) {
        if (intent.action == Intent.ACTION_BOOT_COMPLETED) {
           startStickyNotification(context)
        } else {
            Timber.e("invalid intent [${intent.action}]")
        }
    }
    private fun startStickyNotification(context: Context) {
        if(!Settings.getActiveStatus(context, Constants.SIM1) && !Settings.getActiveStatus(context, Constants.SIM2)) {
            Timber.d("active status is false, not starting foreground service")
            return
        }

        Timber.d("starting foreground service")
        val notificationIntent = Intent(context, StickyNotificationService::class.java)
        val service = context.startForegroundService(notificationIntent)
        Timber.d("foreground service started [${service?.className}]")
    }
}
