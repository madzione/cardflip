import 'package:flutter/material.dart';
import 'package:flutter/semantics.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/l10n.dart';

import 'achievement_badge.dart';
import 'achievements.dart';
import 'session_rewards.dart';

/// The text of a reward banner for [events]: "Daily goal reached · Level 13
/// · New: 1,000 answers".
String rewardBannerText(AppLocalizations l, RewardEvents events) {
  final named = events.named;
  return [
    if (events.goal) l.goalReached,
    if (events.level case final level?) l.levelBanner(level),
    if (named.isNotEmpty) newAchievementsLine(l, named),
  ].join(' · ');
}

/// The study screen's banner for what an answer earned (progress spec,
/// "Reward banner during study" and "Reward banner does not interrupt"):
/// a pill with the first tier's badge, or an icon for the level or goal,
/// and the text. It slides up and fades in, stays, and fades out, about
/// 2.5 seconds in all, then calls [onDone]; under reduce motion it appears
/// and disappears at once. It takes no focus and ignores pointers, so keys
/// and taps go to the screen, and screen readers hear its text as a polite
/// announcement instead of finding it. Give a new banner a new key.
class RewardBanner extends StatefulWidget {
  const RewardBanner({
    super.key,
    required this.events,
    this.minimalist = false,
    this.onDone,
  });

  final RewardEvents events;
  final bool minimalist;
  final VoidCallback? onDone;

  static const duration = Duration(milliseconds: 2500);

  @override
  State<RewardBanner> createState() => _RewardBannerState();
}

class _RewardBannerState extends State<RewardBanner>
    with SingleTickerProviderStateMixin {
  // The hold counts under reduce motion too, which would shorten it.
  late final _controller = AnimationController(
    vsync: this,
    duration: RewardBanner.duration,
    animationBehavior: AnimationBehavior.preserve,
  );

  // In over 200 ms, out over the last 300 ms.
  static const _inEnd = 200 / 2500;
  static const _outStart = 2200 / 2500;

  bool _announced = false;

  @override
  void initState() {
    super.initState();
    _controller.forward().whenCompleteOrCancel(() {
      if (mounted && _controller.isCompleted) widget.onDone?.call();
    });
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (_announced) return;
    _announced = true;
    final l = context.l10n;
    SemanticsService.sendAnnouncement(
      View.of(context),
      rewardBannerText(l, widget.events),
      Directionality.of(context),
      assertiveness: Assertiveness.polite,
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final pill = _Pill(events: widget.events, minimalist: widget.minimalist);
    return IgnorePointer(
      child: ExcludeFocus(
        child: ExcludeSemantics(
          child: reduceMotion(context)
              ? pill
              : AnimatedBuilder(
                  key: const Key('reward-banner-motion'),
                  animation: _controller,
                  builder: (context, child) {
                    final t = _controller.value;
                    final shown = t < _outStart
                        ? Curves.easeOut.transform((t / _inEnd).clamp(0, 1))
                        : 1 - (t - _outStart) / (1 - _outStart);
                    final rise = t < _outStart ? 1 - shown : 0.0;
                    return Opacity(
                      opacity: shown.clamp(0, 1),
                      child: Transform.translate(
                        offset: Offset(0, 16 * rise),
                        child: child,
                      ),
                    );
                  },
                  child: pill,
                ),
        ),
      ),
    );
  }
}

class _Pill extends StatelessWidget {
  const _Pill({required this.events, required this.minimalist});

  final RewardEvents events;
  final bool minimalist;

  @override
  Widget build(BuildContext context) {
    final l = context.l10n;
    final theme = Theme.of(context);
    final colors = theme.colorScheme;
    final named = events.named;
    final golang.NewAchievement? first = named.isEmpty ? null : named.first;
    final Widget leading = first != null
        ? AchievementBadge(
            id: first.id,
            tier: first.tierName,
            stars: newTierStars(first),
            size: 28,
            minimalist: minimalist,
          )
        : Icon(
            events.level != null ? Icons.trending_up : Icons.flag_outlined,
            size: 22,
            color: colors.primary,
          );
    return Material(
      key: const Key('reward-banner'),
      color: colors.surfaceContainerHigh,
      elevation: minimalist ? 0 : 3,
      shadowColor: minimalist ? Colors.transparent : null,
      shape: StadiumBorder(side: BorderSide(color: colors.outlineVariant)),
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 480),
        child: Padding(
          padding: const EdgeInsetsDirectional.fromSTEB(10, 6, 18, 6),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              leading,
              const SizedBox(width: 10),
              Flexible(
                child: Text(
                  rewardBannerText(l, events),
                  style: theme.textTheme.bodyMedium,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
