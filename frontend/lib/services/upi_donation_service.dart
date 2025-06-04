// lib/services/upi_donation_service.dart

import 'package:flutter_riverpod/flutter_riverpod.dart';

// UPI Donation Service Provider
final upiDonationServiceProvider = Provider<UpiDonationService>((ref) {
  return UpiDonationService();
});

class UpiDonationService {
  // ⚠️ REPLACE WITH YOUR ACTUAL UPI ID
  static const String _upiId =
      'aditya.nittala5@oksbi'; // Replace with your UPI ID
  static const String _payeeName = 'Aditya Nittala'; // Replace with your name
  static const String _transactionNote = 'Support GoNews Development';

  // Predefined donation amounts
  static const List<DonationAmount> donationAmounts = [
    DonationAmount(
      amount: 50,
      label: '₹50',
      description: 'Buy me a coffee ☕',
      emoji: '☕',
      isCustom: false,
    ),
    DonationAmount(
      amount: 100,
      label: '₹100',
      description: 'Support development 💻',
      emoji: '💻',
      isCustom: false,
    ),
    DonationAmount(
      amount: 200,
      label: '₹200',
      description: 'Boost our servers 🚀',
      emoji: '🚀',
      isCustom: false,
    ),
    DonationAmount(
      amount: 500,
      label: '₹500',
      description: 'Premium supporter 💎',
      emoji: '💎',
      isCustom: false,
    ),
    DonationAmount(
      amount: 1000,
      label: '₹1000',
      description: 'Super supporter 🌟',
      emoji: '🌟',
      isCustom: false,
    ),
    DonationAmount(
      amount: -1, // Special value to indicate custom amount
      label: 'Custom',
      description: 'Enter your amount ✏️',
      emoji: '✏️',
      isCustom: true, // This is the custom option
    ),
  ];

  /// Generate UPI QR code data for a specific amount
  String generateUpiQrData(double amount) {
    final uriData = Uri(
      scheme: 'upi',
      host: 'pay',
      queryParameters: {
        'pa': _upiId, // Virtual Payment Address
        'pn': _payeeName, // Payee Name
        'am': amount.toStringAsFixed(2), // Amount
        'tn': _transactionNote, // Transaction Note
        'cu': 'INR', // Currency
      },
    );

    return uriData.toString();
  }

  /// Generate custom UPI QR code data with custom note
  String generateCustomUpiQrData(double amount, String customNote) {
    final uriData = Uri(
      scheme: 'upi',
      host: 'pay',
      queryParameters: {
        'pa': _upiId,
        'pn': _payeeName,
        'am': amount.toStringAsFixed(2),
        'tn': customNote.isNotEmpty ? customNote : _transactionNote,
        'cu': 'INR',
      },
    );

    return uriData.toString();
  }

  /// Get UPI ID for display
  String get upiId => _upiId;

  /// Get payee name
  String get payeeName => _payeeName;

  /// Validate UPI ID format
  bool isValidUpiId(String upiId) {
    // Basic UPI ID validation
    final upiRegex = RegExp(r'^[a-zA-Z0-9._-]+@[a-zA-Z0-9.-]+$');
    return upiRegex.hasMatch(upiId);
  }
}

// Donation amount model with isCustom property
class DonationAmount {
  final double amount;
  final String label;
  final String description;
  final String emoji;
  final bool isCustom; // This property was missing in your file

  const DonationAmount({
    required this.amount,
    required this.label,
    required this.description,
    required this.emoji,
    this.isCustom = false, // Default to false for regular amounts
  });
}

// UPI app information
class UpiApp {
  final String name;
  final String packageName;
  final String iconAsset;

  const UpiApp({
    required this.name,
    required this.packageName,
    required this.iconAsset,
  });
}

// Popular UPI apps
class UpiApps {
  static const List<UpiApp> popularApps = [
    UpiApp(
      name: 'Google Pay',
      packageName: 'com.google.android.apps.nbu.paisa.user',
      iconAsset: '💳',
    ),
    UpiApp(
      name: 'PhonePe',
      packageName: 'com.phonepe.app',
      iconAsset: '📱',
    ),
    UpiApp(
      name: 'Paytm',
      packageName: 'net.one97.paytm',
      iconAsset: '💰',
    ),
    UpiApp(
      name: 'BHIM',
      packageName: 'in.org.npci.upiapp',
      iconAsset: '🏛️',
    ),
    UpiApp(
      name: 'Amazon Pay',
      packageName: 'in.amazon.mShop.android.shopping',
      iconAsset: '📦',
    ),
  ];
}
