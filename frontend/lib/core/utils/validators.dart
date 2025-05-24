// lib/core/utils/validators.dart

import '../constants/app_constants.dart';

class Validators {
  // Email validation
  static String? email(String? value) {
    if (value == null || value.isEmpty) {
      return 'Email is required';
    }

    final emailRegex = RegExp(AppConstants.emailPattern);
    if (!emailRegex.hasMatch(value)) {
      return 'Please enter a valid email address';
    }

    return null;
  }

  // Password validation
  static String? password(String? value) {
    if (value == null || value.isEmpty) {
      return 'Password is required';
    }

    if (value.length < 6) {
      return 'Password must be at least 6 characters long';
    }

    return null;
  }

  // Strong password validation (for sign up)
  static String? strongPassword(String? value) {
    if (value == null || value.isEmpty) {
      return 'Password is required';
    }

    if (value.length < 8) {
      return 'Password must be at least 8 characters long';
    }

    // Check for at least one uppercase letter
    if (!value.contains(RegExp(r'[A-Z]'))) {
      return 'Password must contain at least one uppercase letter';
    }

    // Check for at least one lowercase letter
    if (!value.contains(RegExp(r'[a-z]'))) {
      return 'Password must contain at least one lowercase letter';
    }

    // Check for at least one digit
    if (!value.contains(RegExp(r'[0-9]'))) {
      return 'Password must contain at least one number';
    }

    return null;
  }

  // Confirm password validation
  static String? confirmPassword(String? value, String originalPassword) {
    if (value == null || value.isEmpty) {
      return 'Please confirm your password';
    }

    if (value != originalPassword) {
      return 'Passwords do not match';
    }

    return null;
  }

  // Name validation
  static String? name(String? value) {
    if (value == null || value.isEmpty) {
      return 'Name is required';
    }

    if (value.length < 2) {
      return 'Name must be at least 2 characters long';
    }

    if (value.length > 50) {
      return 'Name must be less than 50 characters';
    }

    // Check for valid name characters (letters, spaces, hyphens, apostrophes)
    final nameRegex = RegExp(r"^[a-zA-Z\s\-'\.]+$");
    if (!nameRegex.hasMatch(value)) {
      return 'Name can only contain letters, spaces, hyphens, and apostrophes';
    }

    return null;
  }

  // Phone number validation (Indian format)
  static String? phoneNumber(String? value) {
    if (value == null || value.isEmpty) {
      return 'Phone number is required';
    }

    final phoneRegex = RegExp(AppConstants.phonePattern);
    if (!phoneRegex.hasMatch(value)) {
      return 'Please enter a valid Indian mobile number';
    }

    return null;
  }

  // Optional phone number validation
  static String? optionalPhoneNumber(String? value) {
    if (value == null || value.isEmpty) {
      return null; // Optional field
    }

    return phoneNumber(value);
  }

  // Required field validation
  static String? required(String? value, String fieldName) {
    if (value == null || value.trim().isEmpty) {
      return '$fieldName is required';
    }
    return null;
  }

  // URL validation
  static String? url(String? value) {
    if (value == null || value.isEmpty) {
      return 'URL is required';
    }

    final urlRegex = RegExp(
        r'^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)$');

    if (!urlRegex.hasMatch(value)) {
      return 'Please enter a valid URL';
    }

    return null;
  }

  // Age validation
  static String? age(String? value) {
    if (value == null || value.isEmpty) {
      return 'Age is required';
    }

    final age = int.tryParse(value);
    if (age == null) {
      return 'Please enter a valid age';
    }

    if (age < 13) {
      return 'You must be at least 13 years old';
    }

    if (age > 120) {
      return 'Please enter a valid age';
    }

    return null;
  }

  // OTP validation
  static String? otp(String? value) {
    if (value == null || value.isEmpty) {
      return 'OTP is required';
    }

    if (value.length != 6) {
      return 'OTP must be 6 digits';
    }

    final otpRegex = RegExp(r'^[0-9]{6}$');
    if (!otpRegex.hasMatch(value)) {
      return 'OTP must contain only numbers';
    }

    return null;
  }

  // Search query validation
  static String? searchQuery(String? value) {
    if (value == null || value.trim().isEmpty) {
      return 'Please enter a search term';
    }

    if (value.trim().length < 2) {
      return 'Search term must be at least 2 characters';
    }

    return null;
  }

  // Generic length validation
  static String? length(String? value, int min, int max, String fieldName) {
    if (value == null || value.isEmpty) {
      return '$fieldName is required';
    }

    if (value.length < min) {
      return '$fieldName must be at least $min characters long';
    }

    if (value.length > max) {
      return '$fieldName must be less than $max characters';
    }

    return null;
  }
}
