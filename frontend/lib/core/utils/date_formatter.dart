// lib/core/utils/date_formatter.dart

import 'package:intl/intl.dart';

class DateFormatter {
  // IST Timezone offset
  static const Duration _istOffset = Duration(hours: 5, minutes: 30);

  // Date formatters
  static final DateFormat _timeFormat = DateFormat('h:mm a');
  static final DateFormat _dateFormat = DateFormat('dd MMM yyyy');
  static final DateFormat _dateTimeFormat = DateFormat('dd MMM yyyy, h:mm a');
  static final DateFormat _fullDateFormat = DateFormat('EEEE, dd MMMM yyyy');
  static final DateFormat _monthYearFormat = DateFormat('MMMM yyyy');
  static final DateFormat _dayMonthFormat = DateFormat('dd MMM');

  /// Convert UTC DateTime to IST
  static DateTime toIST(DateTime utcDateTime) {
    return utcDateTime.add(_istOffset);
  }

  /// Get current IST time
  static DateTime get currentIST {
    return DateTime.now().toUtc().add(_istOffset);
  }

  /// Format DateTime to IST time string (e.g., "2:30 PM")
  static String formatTimeIST(DateTime dateTime) {
    final istTime = toIST(dateTime);
    return _timeFormat.format(istTime);
  }

  /// Format DateTime to IST date string (e.g., "24 May 2024")
  static String formatDateIST(DateTime dateTime) {
    final istTime = toIST(dateTime);
    return _dateFormat.format(istTime);
  }

  /// Format DateTime to IST date and time string (e.g., "24 May 2024, 2:30 PM")
  static String formatDateTimeIST(DateTime dateTime) {
    final istTime = toIST(dateTime);
    return _dateTimeFormat.format(istTime);
  }

  /// Format DateTime to full date format (e.g., "Friday, 24 May 2024")
  static String formatFullDateIST(DateTime dateTime) {
    final istTime = toIST(dateTime);
    return _fullDateFormat.format(istTime);
  }

  /// Format DateTime to relative time (e.g., "2 hours ago", "Just now")
  static String formatRelativeTime(DateTime dateTime) {
    final now = currentIST;
    final istTime = toIST(dateTime);
    final difference = now.difference(istTime);

    if (difference.inMinutes < 1) {
      return 'Just now';
    } else if (difference.inMinutes < 60) {
      return '${difference.inMinutes}m ago';
    } else if (difference.inHours < 24) {
      return '${difference.inHours}h ago';
    } else if (difference.inDays < 7) {
      return '${difference.inDays}d ago';
    } else if (difference.inDays < 30) {
      final weeks = (difference.inDays / 7).floor();
      return '${weeks}w ago';
    } else if (difference.inDays < 365) {
      final months = (difference.inDays / 30).floor();
      return '${months}mo ago';
    } else {
      final years = (difference.inDays / 365).floor();
      return '${years}y ago';
    }
  }

  /// Format DateTime to news article timestamp format
  /// Today: "2:30 PM"
  /// Yesterday: "Yesterday, 2:30 PM"
  /// This week: "Monday, 2:30 PM"
  /// Older: "24 May 2024"
  static String formatToIST(DateTime dateTime) {
    final now = currentIST;
    final istTime = toIST(dateTime);
    final difference = now.difference(istTime);

    if (difference.inDays == 0) {
      // Today - show time only
      return formatTimeIST(dateTime);
    } else if (difference.inDays == 1) {
      // Yesterday
      return 'Yesterday, ${formatTimeIST(dateTime)}';
    } else if (difference.inDays < 7) {
      // This week - show day and time
      final dayFormat = DateFormat('EEEE');
      return '${dayFormat.format(istTime)}, ${formatTimeIST(dateTime)}';
    } else {
      // Older - show date only
      return formatDateIST(dateTime);
    }
  }

  /// Check if given time is within Indian market hours (9:15 AM - 3:30 PM IST)
  static bool isMarketHours(DateTime dateTime) {
    final istTime = toIST(dateTime);
    final weekday = istTime.weekday;

    // Market is closed on weekends
    if (weekday == DateTime.saturday || weekday == DateTime.sunday) {
      return false;
    }

    final hour = istTime.hour;
    final minute = istTime.minute;
    final totalMinutes = hour * 60 + minute;

    // Market hours: 9:15 AM to 3:30 PM IST
    const marketOpenMinutes = 9 * 60 + 15; // 9:15 AM
    const marketCloseMinutes = 15 * 60 + 30; // 3:30 PM

    return totalMinutes >= marketOpenMinutes &&
        totalMinutes <= marketCloseMinutes;
  }

  /// Check if given time is within IPL match hours (3:00 PM - 11:00 PM IST)
  static bool isIPLTime(DateTime dateTime) {
    final istTime = toIST(dateTime);
    final hour = istTime.hour;

    // IPL matches typically run from 3:00 PM to 11:00 PM IST
    return hour >= 15 && hour <= 23;
  }

  /// Check if given time is within Indian business hours (9:00 AM - 6:00 PM IST)
  static bool isBusinessHours(DateTime dateTime) {
    final istTime = toIST(dateTime);
    final weekday = istTime.weekday;

    // Business hours are Monday to Friday
    if (weekday == DateTime.saturday || weekday == DateTime.sunday) {
      return false;
    }

    final hour = istTime.hour;
    return hour >= 9 && hour <= 18;
  }

  /// Format duration in a human-readable format
  static String formatDuration(Duration duration) {
    if (duration.inDays > 0) {
      return '${duration.inDays}d ${duration.inHours % 24}h';
    } else if (duration.inHours > 0) {
      return '${duration.inHours}h ${duration.inMinutes % 60}m';
    } else if (duration.inMinutes > 0) {
      return '${duration.inMinutes}m';
    } else {
      return '${duration.inSeconds}s';
    }
  }

  /// Get greeting based on IST time
  static String getGreeting() {
    final istTime = currentIST;
    final hour = istTime.hour;

    if (hour >= 5 && hour < 12) {
      return 'Good Morning';
    } else if (hour >= 12 && hour < 17) {
      return 'Good Afternoon';
    } else if (hour >= 17 && hour < 21) {
      return 'Good Evening';
    } else {
      return 'Good Night';
    }
  }

  /// Get time ago in a human-readable format
  static String getTimeAgo(DateTime dateTime) {
    final Duration difference = DateTime.now().difference(dateTime);

    if (difference.inDays > 1) {
      return '${difference.inDays} days ago';
    } else if (difference.inHours > 1) {
      return '${difference.inHours} hours ago';
    } else if (difference.inMinutes > 1) {
      return '${difference.inMinutes} minutes ago';
    } else {
      return 'Just now';
    }
  }

  /// Parse ISO string to DateTime
  static DateTime? parseISOString(String? isoString) {
    if (isoString == null || isoString.isEmpty) return null;
    try {
      return DateTime.parse(isoString);
    } catch (e) {
      return null;
    }
  }
}
