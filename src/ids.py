import time

import requests
import re
import random
import string
from Crypto.Cipher import AES
from Crypto.Util.Padding import pad
from Crypto.Random import get_random_bytes
import base64
from bs4 import BeautifulSoup
import json


class XMULogin:
    def __init__(self):
        self.session = requests.Session()
        # 设置User-Agent模拟真实浏览器
        self.session.headers.update({
            'User-Agent': 'Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Mobile Safari/537.36'
        })

    def get_random_string(self, length):
        """生成随机字符串，模拟JavaScript中的randomString函数"""
        chars = "ABCDEFGHJKMNPQRSTWXYZabcdefhijkmnprstwxyz2345678"
        return ''.join(random.choice(chars) for _ in range(length))

    def aes_encrypt(self, plaintext, key, iv):
        """AES加密函数"""
        # 将key和iv转换为字节
        key_bytes = key.encode('utf-8')
        iv_bytes = iv.encode('utf-8')
        plaintext_bytes = plaintext.encode('utf-8')

        # 创建AES加密器（CBC模式）
        cipher = AES.new(key_bytes, AES.MODE_CBC, iv_bytes)

        # 填充并加密
        padded_text = pad(plaintext_bytes, AES.block_size)
        encrypted = cipher.encrypt(padded_text)

        # 返回base64编码的结果
        return base64.b64encode(encrypted).decode('utf-8')

    def encrypt_password(self, password, salt):
        """模拟JavaScript中的encryptPassword函数"""
        if not salt:
            return password

        try:
            # 生成64位随机字符串 + 密码
            random_prefix = self.get_random_string(64)
            combined_text = random_prefix + password

            # 生成16位随机IV
            iv = self.get_random_string(16)

            # 去除salt的前后空格
            salt = salt.strip()

            # AES加密
            encrypted = self.aes_encrypt(combined_text, salt, iv)
            return encrypted

        except Exception as e:
            print(f"加密过程中出现错误: {e}")
            return password

    def get_login_page(self):
        """获取登录页面并提取必要信息"""
        url = "https://ids.xmu.edu.cn/authserver/login"

        try:
            response = self.session.get(url)
            response.raise_for_status()

            # 解析HTML
            soup = BeautifulSoup(response.text, 'html.parser')

            # 提取加密盐
            salt_element = soup.find('input', {'id': 'pwdEncryptSalt'})
            salt = salt_element.get('value') if salt_element else None

            # 提取execution值
            execution_element = soup.find('input', {'name': 'execution'})
            execution = execution_element.get('value') if execution_element else None

            # 提取lt值（如果存在）
            lt_element = soup.find('input', {'name': 'lt'})
            lt = lt_element.get('value') if lt_element else ""

            print(f"成功获取登录页面")
            print(f"加密盐: {salt}")
            print(f"Execution: {execution}")
            print(f"LT: {lt}")

            return salt, execution, lt

        except requests.RequestException as e:
            print(f"获取登录页面失败: {e}")
            return None, None, None

    def login(self, username, password):
        """执行登录操作"""
        # 获取登录页面信息
        salt, execution, lt = self.get_login_page()

        if not salt or not execution:
            print("无法获取必要的登录信息")
            return False

        # 加密密码
        encrypted_password = self.encrypt_password(password, salt)
        print(f"加密后的密码: {encrypted_password}")

        # 准备登录数据
        login_data = {
            'username': username,
            'password': encrypted_password,
            'captcha': '',
            '_eventId': 'submit',
            'lt': lt,
            'cllt': 'userNameLogin',
            'dllt': 'generalLogin',
            'execution': execution
        }

        # 设置登录请求头
        headers = {
            'Content-Type': 'application/x-www-form-urlencoded',
            'Referer': 'https://ids.xmu.edu.cn/authserver/login'
        }

        try:
            # 发送登录请求
            login_url = "https://ids.xmu.edu.cn/authserver/login"
            response = self.session.post(login_url, data=login_data, headers=headers, allow_redirects=False)

            print(f"登录响应状态码: {response.status_code}")

            # 检查登录结果
            if response.status_code == 302:
                # 重定向通常表示登录成功
                location = response.headers.get('Location', '')
                print(f"登录成功，重定向到: {location}")
                return True
            elif "用户名或密码错误" in response.text or "errorMessage" in response.text:
                print("登录失败：用户名或密码错误")
                return False
            else:
                print(f"登录状态未知，响应内容: {response.text[:500]}")
                return False

        except requests.RequestException as e:
            print(f"登录请求失败: {e}")
            return False

    def rollCallStatus(self):
        """查询签到状态，返回需要签到的课程信息"""
        url = "https://lnt.xmu.edu.cn/api/radar/rollcalls?api_version=1.1.0"

        try:
            response = self.session.get(url)
            response.raise_for_status()

            print(f"签到状态查询响应: {response.text}")

            # 解析JSON响应
            data = response.json()
            rollcalls = data.get('rollcalls', [])

            # 提取需要签到的课程信息
            pending_rollcalls = {}

            for rollcall in rollcalls:
                # 检查是否需要签到（状态为in_progress且学生状态为absent）
                if (rollcall.get('rollcall_status') == 'in_progress' and
                        rollcall.get('status') == 'absent' and
                        not rollcall.get('is_expired', True)):

                    course_title = rollcall.get('course_title', '')
                    rollcall_id = rollcall.get('rollcall_id')

                    if rollcall_id:
                        pending_rollcalls[course_title] = rollcall_id
                        print(f"需要签到的课程: {course_title} (ID: {rollcall_id})")

            if not pending_rollcalls:
                print("当前没有需要签到的课程")

            return pending_rollcalls

        except requests.RequestException as e:
            print(f"查询签到状态失败: {e}")
            return {}
        except json.JSONDecodeError as e:
            print(f"解析JSON响应失败: {e}")
            return {}

    def rollCallAnswer(self, rollcall_dict):
        """执行签到操作"""
        if not rollcall_dict:
            print("没有需要签到的课程")
            return {}

        rollcall_codes = {}

        for course_title, rollcall_id in rollcall_dict.items():
            url = f"https://lnt.xmu.edu.cn/api/rollcall/{rollcall_id}/student_rollcalls"

            try:
                print(f"正在查询课程 '{course_title}' 的签到码...")
                response = self.session.get(url)
                response.raise_for_status()

                print(f"签到码查询响应: {response.text}")

                # 解析JSON响应
                data = response.json()
                number_code = data.get('number_code')

                if number_code:
                    rollcall_codes[course_title] = number_code
                    print(f"课程 '{course_title}' 的签到码: {number_code}")
                else:
                    print(f"课程 '{course_title}' 未找到签到码")
                    rollcall_codes[course_title] = None

            except requests.RequestException as e:
                print(f"查询课程 '{course_title}' 签到码失败: {e}")
                rollcall_codes[course_title] = None
            except json.JSONDecodeError as e:
                print(f"解析课程 '{course_title}' 响应失败: {e}")
                rollcall_codes[course_title] = None

        return rollcall_codes

    def rollCallAnswerTest(self, id):

            url = f"https://lnt.xmu.edu.cn/api/rollcall/{id}/student_rollcalls"


            print(f"正在查询课程签到码...")
            response = self.session.get(url)
            response.raise_for_status()

            print(f"签到码查询响应: {response.status_code}")

            # 解析JSON响应
            data = response.json()
            number_code = data.get('number_code')
            print(f"签到码: {number_code}")

def main():
    # 创建登录实例
    xmu_login = XMULogin()

    # 登录信息
    username = None
    password = None

    with open("cfg.txt", "r", encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if line.startswith("username:"):
                # 去掉前缀并去掉引号
                username = line.split(":", 1)[1].strip().strip('"')
            elif line.startswith("password:"):
                password = line.split(":", 1)[1].strip().strip('"')

    print("username =", username)
    print("password =", password)

    print("开始模拟登录厦门大学统一认证系统...")

    # 执行登录
    if xmu_login.login(username, password):
        print("登录成功！")
        # xmu_login.rollCallAnswerTest(141798)
        while(True):
            # 查询签到状态
            print("\n=== 查询签到状态 ===")
            pending_rollcalls = xmu_login.rollCallStatus()

            if pending_rollcalls:
                # 执行签到
                print("\n=== 签到码查询操作 ===")
                rollcall_codes = xmu_login.rollCallAnswer(pending_rollcalls)

                # 打印结果总结
                print("\n=== 签到结果总结 ===")
                for course_title, code in rollcall_codes.items():
                    if code:
                        print(f"✅ {course_title}: 签到码 {code}")
                    else:
                        print(f"❌ {course_title}: 获取签到码失败")
                time.sleep(200)
                break
            else:
                time.sleep(2)
    else:
        print("登录失败，无法进行签到操作")


if __name__ == "__main__":
    main()
