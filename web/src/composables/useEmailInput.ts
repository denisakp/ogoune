import { computed, type Ref } from 'vue'

export function useEmailInput(emailArray: Ref<string[] | undefined>) {
  console.log('useEmailInput initialized with:', emailArray.value)
  return computed({
    get: () => (Array.isArray(emailArray.value) ? emailArray.value.join(', ') : ' '),
    set: (value: string) => {
      emailArray.value = value
        .split(',')
        .map((email) => email.trim())
        .filter((email) => email.length > 0)
    },
  })
}

export function useEmailValidation() {
  const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/

  const validateRmailArray = (allowEmpty = true) => {
    return (_rule: any, value: string[]) => {
      if (!value || value.length === 0) {
        return allowEmpty
          ? Promise.resolve()
          : Promise.reject(new Error('At least one email is required.'))
      }

      const invalidEmails = value.filter((email) => !emailRegex.test(email))

      if (invalidEmails.length > 0)
        return Promise.reject(new Error(`Invalid email(s): ${invalidEmails.join(', ')}`))

      return Promise.resolve()
    }
  }

  const validateEmailString = () => {
    return (_rule: any, value: string) => {
      if (!value) return Promise.resolve()

      if (!emailRegex.test(value)) return Promise.reject(new Error('Invalid email address.'))

      return Promise.resolve()
    }
  }

  const isValidemail = (email: string): boolean => {
    return emailRegex.test(email)
  }

  return {
    validateRmailArray,
    validateEmailString,
    isValidemail,
  }
}
