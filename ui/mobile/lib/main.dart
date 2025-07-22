import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/material.dart';
import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_auth/firebase_auth.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';
import 'firebase_options.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await Firebase.initializeApp(options: DefaultFirebaseOptions.currentPlatform);
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: '2FA Sender',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.amber),
      ),
      home: const MyHomePage(title: '2FA Trigger Page'),
    );
  }
}

class MyHomePage extends StatefulWidget {
  const MyHomePage({super.key, required this.title});
  final String title;

  @override
  State<MyHomePage> createState() => _MyHomePageState();
}

class _MyHomePageState extends State<MyHomePage> {
  String status = 'Нажми кнопку для отправки';

  @override
  void initState() {
    super.initState();

    // Обработка уведомлений при получении в foreground
    FirebaseMessaging.onMessage.listen((RemoteMessage message) {
      print('📥 Получено сообщение в foreground!');
      print('🔹 Title: ${message.notification?.title}');
      print('🔹 Body: ${message.notification?.body}');

      setState(() {
        status = 'Push получен в foreground: ${message.notification?.title}';
      });
    });

    // Обработка, если пользователь тапнул по уведомлению и приложение в background
    FirebaseMessaging.onMessageOpenedApp.listen((RemoteMessage message) {
      print('📬 Пользователь открыл уведомление');
      print('🔹 Title: ${message.notification?.title}');
      print('🔹 Body: ${message.notification?.body}');

      setState(() {
        status = 'Push открыт из background: ${message.notification?.title}';
      });
    });

    // Обработка пуша при запуске приложения
    FirebaseMessaging.instance.getInitialMessage().then((message) {
      if (message != null) {
        print('🚀 Приложение открыто по пушу');
        setState(() {
          status = 'Push при запуске: ${message.notification?.title}';
        });
      }
    });
  }

  Future<void> _send2FARequest() async {
    try {
      FirebaseMessaging messaging = FirebaseMessaging.instance;

      NotificationSettings settings = await messaging.requestPermission(
        alert: true,
        badge: true,
        sound: true,
      );

      print('User granted permission: ${settings.authorizationStatus}');

      final token = await FirebaseMessaging.instance.getToken();
      print('📱 FCM Token: $token');
    } catch (e) {
      setState(() {
        status = '⚠️ Ошибка: $e';
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Theme.of(context).colorScheme.inverseSurface,
        title: Text(widget.title),
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: <Widget>[
            Text(status, textAlign: TextAlign.center),
            const SizedBox(height: 20),
            ElevatedButton(
              onPressed: _send2FARequest,
              child: const Text('Отправить 2FA-запрос'),
            ),
          ],
        ),
      ),
    );
  }
}
