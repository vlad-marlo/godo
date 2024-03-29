# 1. Цель проекта

Цель проекта - разработать систему для организации работы в команде (далее Система). Пользователь сможет создавать
группы, задачи. Для групп будет предусмотрена внутренняя система прав(создание, чтение, изменение задач, группы).
Помимо этого в Системе должна быть предусмотрена интеграция с Telegram чатом, в котором будут находиться только участники группы.


# 2. Описание системы

Система состоит из следующих основных функциональных блоков:
1. Регистрация, аутентификация и авторизация
2. Функционал для администратора группы
3. Функционал для обычного пользователя
4. Функционал интеграции с Telegram
5. Уведомления о новых задачах


## 2.1. Типы пользователей

Система предусматривает два типа пользователей внутри групп: администратор и обычный пользователь.
Внутри группы существует два типа пользователей: администраторы и обычные пользователи. Изначально у администраторов
есть права на чтение, создание, изменение, удаление задач внутри группы, когда как у обычных пользователей
только на чтение, изменение статуса задачи (см пункт 2.4.х) и запрос проверки (ревью), создание проблемы, связанной с задачей
(как Issue на github).

Для каждой графы роли уместны соответствуют права
* 0 - прав нет;
* 1 - права на чтение;
* 2 - права на чтение, запись, изменение своих записей;
* 3 - права на чтение, запись и изменение чужих записей;
* 4 - любые права;


## 2.2. Регистрация пользователей

Процесс заведения администратора в инсталляции Системы может быть реализован
в первой версии не в интерфейсе Системы, а при помощи команды на сервере,
это допускается. Этой команде заведения автора должны быть переданы на вход
следующие данные:

* email — обязательное поле
* пароль — обязательное поле

Процесс заведения автора через команду на сервере должен быть описан в
документации Системы.

Процесс регистрации подписчиков, конечно, должен быть реализован в
интерфейсах Системы. При регистрации подписчика должны быть запрошены
следующие поля:

* email — обязательное поле
* имя и фамилия — обязательное поле
* о себе - опциональное поле
* пароль — обязательное поле


## 2.3. Аутентификация

Аутентификация пользователя осуществляется по Bearer токену, который можно получить в интерфейсе Системы по имени пользователя и
паролю.


## 2.4. Функционал для администратора группы

1. Редактирование профиля группы
2. Создание и изменение задач
3. Система проверки (опционально)
4. Создание кастомных прав в группе
5. Изменение группы прав пользователя в группе
6. Система приглашений пользователей в группу
7. Удаление пользователей из группы


### 2.4.1 Редактирование профиля группы

В этом разделе у пользователя есть возможность редактирования данных профиля группы — телеграм чат, название компании,
информация о компании, дополнительные поля.


### 2.4.2 Создание и редактирование задач

Каждой задаче задаются следующие параметры:
* название задачи - обязательное поле
* описание задачи - обязательное поле
* срок сдачи - опциональное поле
* идентификаторы пользователей, для которых создается задача - опционально
* группы - опционально, перечисление идентификаторов групп, для которых создается поле

Если не указывать ни пользователей ни группу, то задача создастся только для пользователя, отправившего запрос.
Для задачи будут созданы дополнительные записи - кортеж, связывающий пользователя с задачей.

У задачи также существует поле статуса. Существуют следующие статусы:
* Создано
* В работе
* На проверке
* Ожидаются правки
* Закрыта
* Принята

Любой пользователь может изменить статус с "Создано" на "В работе" и с "В работе" на "На проверке".

Дополнительно в БД записывается информация о пользователе, создавшем задачу. В реализации системы при создании можно
будет просто указать id группы, для которой создается задачи вместо перечисления всех пользователей.

Для того, что-бы создать задачу для другого пользователя, необходимо иметь общие группы в которых у пользователя
есть право на создание задач. Если указать пользователя, с котором нет общих групп, то задача не создастся.

Изначально любой пользователь может изменить статус задачи (взять ее).


### 2.4.3 Проверка задачи

Автоматически проверить выполнение задачи может только администратор группы, или пользователь, создавший ее.
В Системе должна быть предусмотрена возможность получить задачи, ожидающие проверку.

Для проверки задач будет создана дополнительный функционал проблем.

При запросе задач, ожидающих проверки Система должна возвращать множество объектов со следующими полями:
* идентификатор запроса
* задача - идентификатор задачи (первичный ключ, далее ПК)
* дата сдачи
* пользователь - идентификатор пользователя, сдавшего работу (ПК)
* сообщение - дополнительная информация, которую дает пользователь.

Администратор для изменения статуса должен отправить системе следующий запрос:
* идентификатор запроса - обязательное поле
* сообщение - опциональное поле
* статус - обязательное поле

Возможные статусы:
* ожидаются правки
* принято

Данные статусы описывают лишь фактический смысл, подробнее будет описано в документации Системы.


### 2.4.4 Создание кастомных прав

Администратор может создавать дополнительные права для пользователей в группе.
Роль будет иметь следующие параметры
* создание задач
* изменение задач
* удаление задач
* проверка задачи
* просмотр задач, ожидающих проверки
* создание проблемы (комментария, относящегося к задаче)
* просмотр проблем задач
* просмотр пользователей, состоящих в группе
* создание ссылок-приглашений в группу
* удаление пользователей из группы

Роль с параметрами будет создана всегда только один раз (роль с набором прав будет существовать лишь в единственном экземпляре).
В БД будет связная таблица пользователь-роль-группа.


### 2.4.5 Приглашения в группы

Для того что-бы попасть в группу, пользователю будет необходимо пройти по специальной одноразовой ссылке-приглашению.
Создать данную ссылку может любой пользователь с правами на приглашения.
Также в дальнейшем рассматривается создание направленных приглашений пользователям по username/id.


## 2.5 Права для любых пользователей

Каждый пользователь не зависимо от группы прав имеет следующие права:
1. просмотр собственного профиля
2. просмотр связанных с пользователем задач
3. просмотр статуса работ, отправленных на проверку
4. изменение собственного профиля


## 2.6 Предполагаемый минимум функционала для написания дипломной работы

1. Регистрация, аутентификация пользователей
2. Создание групп пользователей
3. Создание и удаление задач
4. Ревью система

# 3 Предполагаемый стек технологий

- go 1.19+;
- go.uber/fx - DI container;
- postgresql - DB;
- tern - migration tool;
- github.com/grpc/grpc-go - grpc server;
Сервер должен предусматривать работу с использованием как http/https, так и grpc.