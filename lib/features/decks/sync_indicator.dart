import 'package:flutter/material.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/app/errors.dart';
import 'package:cardflip/app/scaffold.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/l10n.dart';
import 'package:cardflip/widgets/time_format.dart';

/// The deck list's sync indicator, shown while the open profile has a link
/// (device-sync spec: Sync indicator): syncing, synced with the time, not
/// on a trusted network, needs your choice, or failed. Tapping it syncs
/// now and reports what was received and sent; after a failure it first
/// shows the reason. Automatic syncs that fail show only here. Below the
/// wide layout, it is an icon with the state as its tooltip.
class SyncIndicator extends StatefulWidget {
  const SyncIndicator({super.key});

  @override
  State<SyncIndicator> createState() => _SyncIndicatorState();
}

class _SyncIndicatorState extends State<SyncIndicator> {
  bool _busy = false;

  /// Syncs the open profile with every reachable linked device (device-sync
  /// spec: Sync now).
  Future<void> _syncNow() async {
    setState(() => _busy = true);
    try {
      final sum = await golang.sync.syncNow();
      if (sum.error case final e? when e.isNotEmpty) {
        showError(golang.GolangException('sync.SyncNow', e));
      } else if (!sum.reachable) {
        showInfo(appL10n.syncNothingReachable);
      } else if (sum.media > 0) {
        showInfo(appL10n.syncSummaryMedia(sum.received, sum.sent, sum.media));
      } else {
        showInfo(appL10n.syncSummary(sum.received, sum.sent));
      }
    } catch (e) {
      showError(e);
    }
    if (mounted) setState(() => _busy = false);
  }

  /// Shows why the last sync failed and offers to sync again.
  Future<void> _showError(String error) async {
    final again = await showDialog<bool>(
      context: context,
      animationStyle: popupAnimation(context),
      builder: (context) => AlertDialog(
        icon: const Icon(Icons.sync_problem),
        title: Text(context.l10n.syncFailed),
        content: Text(
          errorMessage(golang.GolangException('sync', error)),
          key: const Key('sync-error'),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(context.l10n.close),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text(context.l10n.syncAgain),
          ),
        ],
      ),
    );
    if ((again ?? false) && mounted) await _syncNow();
  }

  @override
  Widget build(BuildContext context) {
    final sync = AppScope.of(context).sync;
    return ListenableBuilder(
      listenable: sync,
      builder: (context, _) {
        final st = sync.status;
        if (!sync.supported || st == null || st.links.isEmpty) {
          return const SizedBox.shrink();
        }
        final l = context.l10n;
        final scheme = Theme.of(context).colorScheme;
        final error = [
          st.error,
          for (final link in st.links) link.error,
        ].firstWhere((e) => e != null && e.isNotEmpty, orElse: () => null);
        DateTime? last = st.lastSynced;
        for (final link in st.links) {
          if (link.lastSynced case final t?
              when last == null || t.isAfter(last)) {
            last = t;
          }
        }
        final IconData icon;
        final String label;
        Color? color;
        VoidCallback? onTap = _syncNow;
        if (st.syncing || _busy) {
          icon = Icons.sync;
          label = switch (st.progress) {
            final p? => l.syncProgress(p.received, p.sent),
            null => l.syncSyncing,
          };
          onTap = null;
        } else if (!st.network.trusted) {
          icon = Icons.sync_disabled;
          label = l.syncNotTrusted;
        } else if (st.links.any((link) => link.needsChoice)) {
          icon = Icons.sync_problem;
          label = l.syncNeedsChoice;
          color = scheme.tertiary;
        } else if (error != null) {
          icon = Icons.sync_problem;
          label = l.syncFailed;
          color = scheme.error;
          onTap = () => _showError(error);
        } else {
          icon = Icons.sync;
          label = last == null ? l.neverSynced : l.syncedAt(formatRecent(last));
        }
        if (MediaQuery.sizeOf(context).width < wideLayoutWidth) {
          return IconButton(
            key: const Key('sync-indicator'),
            tooltip: label,
            icon: Icon(icon, color: color),
            onPressed: onTap,
          );
        }
        return Tooltip(
          message: l.syncNow,
          child: TextButton.icon(
            key: const Key('sync-indicator'),
            icon: Icon(icon, color: color),
            label: Text(label),
            onPressed: onTap,
          ),
        );
      },
    );
  }
}
