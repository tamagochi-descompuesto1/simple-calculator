import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError, calculate } from '../api/apiClient'
import { Calculator } from './Calculator'

// Replace only the network call; ApiError stays real, since the component
// checks for it.
vi.mock('../api/apiClient', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../api/apiClient')>()
  return { ...actual, calculate: vi.fn() }
})

const calculateMock = vi.mocked(calculate)

beforeEach(() => {
  calculateMock.mockReset()
})

// Small helpers so the tests read as what the user does.
async function selectOperation(
  user: ReturnType<typeof userEvent.setup>,
  name: string,
) {
  await user.click(screen.getByRole('radio', { name }))
}

async function typeNumber(
  user: ReturnType<typeof userEvent.setup>,
  label: string,
  text: string,
) {
  await user.type(screen.getByLabelText(label), text)
}

const calculateButton = () => screen.getByRole('button', { name: 'Calculate' })

describe('Calculator', () => {
  describe('before calculating', () => {
    it('starts with Add selected, empty fields and nothing to submit', () => {
      render(<Calculator />)

      expect(screen.getByRole('radio', { name: 'Add' })).toBeChecked()
      expect(screen.getByLabelText('First number')).toHaveValue('')
      expect(screen.getByLabelText('Second number')).toHaveValue('')
      expect(calculateButton()).toBeDisabled()
      expect(screen.queryByText('Enter a valid number')).not.toBeInTheDocument()
    })

    it('enables the button once both numbers are valid', async () => {
      const user = userEvent.setup()
      render(<Calculator />)

      await typeNumber(user, 'First number', '6')
      expect(calculateButton()).toBeDisabled()

      await typeNumber(user, 'Second number', '3')
      expect(calculateButton()).toBeEnabled()
    })

    it('shows an error for a number that is not valid', async () => {
      const user = userEvent.setup()
      render(<Calculator />)

      await typeNumber(user, 'First number', 'abc')

      expect(screen.getByLabelText('First number')).toBeInvalid()
      expect(screen.getByText('Enter a valid number')).toBeInTheDocument()
      expect(calculateButton()).toBeDisabled()
    })

    it('accepts a decimal point while the number is being typed', async () => {
      const user = userEvent.setup()
      render(<Calculator />)

      await typeNumber(user, 'First number', '5.')

      expect(screen.getByLabelText('First number')).toBeValid()
    })
  })

  describe('operations with one number', () => {
    it('hides the second number for the square root', async () => {
      const user = userEvent.setup()
      render(<Calculator />)

      await selectOperation(user, 'Square Root')

      expect(screen.queryByLabelText('Second number')).not.toBeInTheDocument()
      expect(screen.getByLabelText('Number')).toBeInTheDocument()
    })

    it('needs only one valid number to be enabled', async () => {
      const user = userEvent.setup()
      render(<Calculator />)

      await selectOperation(user, 'Square Root')
      await typeNumber(user, 'Number', '16')

      expect(calculateButton()).toBeEnabled()
    })

    it('sends only one number and shows the result', async () => {
      const user = userEvent.setup()
      calculateMock.mockResolvedValue(4)
      render(<Calculator />)

      await selectOperation(user, 'Square Root')
      await typeNumber(user, 'Number', '16')
      await user.click(calculateButton())

      expect(calculateMock).toHaveBeenCalledWith('sqrt', 16, undefined)
      expect(await screen.findByRole('status')).toHaveTextContent('√16 = 4')
    })
  })

  describe('calculating', () => {
    it('sends the numbers and shows the result', async () => {
      const user = userEvent.setup()
      calculateMock.mockResolvedValue(9)
      render(<Calculator />)

      await typeNumber(user, 'First number', '6')
      await typeNumber(user, 'Second number', '3')
      await user.click(calculateButton())

      expect(calculateMock).toHaveBeenCalledWith('add', 6, 3)
      expect(await screen.findByRole('status')).toHaveTextContent('6 + 3 = 9')
    })

    it('sends the selected operation', async () => {
      const user = userEvent.setup()
      calculateMock.mockResolvedValue(18)
      render(<Calculator />)

      await selectOperation(user, 'Multiply')
      await typeNumber(user, 'First number', '6')
      await typeNumber(user, 'Second number', '3')
      await user.click(calculateButton())

      expect(calculateMock).toHaveBeenCalledWith('multiply', 6, 3)
      expect(await screen.findByRole('status')).toHaveTextContent('6 × 3 = 18')
    })

    it('words a percentage as "b% of a"', async () => {
      const user = userEvent.setup()
      calculateMock.mockResolvedValue(30)
      render(<Calculator />)

      await selectOperation(user, 'Percentage')
      await typeNumber(user, 'First number', '200')
      await typeNumber(user, 'Second number', '15')
      await user.click(calculateButton())

      expect(await screen.findByRole('status')).toHaveTextContent(
        '15% of 200 = 30',
      )
    })

    it('rounds floating-point noise in the result', async () => {
      const user = userEvent.setup()
      calculateMock.mockResolvedValue(0.1 + 0.2)
      render(<Calculator />)

      await typeNumber(user, 'First number', '0.1')
      await typeNumber(user, 'Second number', '0.2')
      await user.click(calculateButton())

      expect(await screen.findByRole('status')).toHaveTextContent(
        '0.1 + 0.2 = 0.3',
      )
    })

    it('calculates when Enter is pressed in a field', async () => {
      const user = userEvent.setup()
      calculateMock.mockResolvedValue(9)
      render(<Calculator />)

      await typeNumber(user, 'First number', '6')
      await typeNumber(user, 'Second number', '3{Enter}')

      expect(calculateMock).toHaveBeenCalledTimes(1)
      expect(await screen.findByRole('status')).toHaveTextContent('6 + 3 = 9')
    })

    // The browser already refuses to submit through a disabled button, so these
    // two tests send the submit event straight to the form to check that the
    // component guards against it too.
    it('does not calculate if submitted with invalid numbers', async () => {
      const user = userEvent.setup()
      const { container } = render(<Calculator />)
      await typeNumber(user, 'First number', '6')

      fireEvent.submit(container.querySelector('form')!)

      expect(calculateMock).not.toHaveBeenCalled()
      // Nothing happened at all, not even an error message.
      expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    })

    it('does not calculate twice if submitted while waiting', async () => {
      const user = userEvent.setup()
      calculateMock.mockReturnValue(new Promise<number>(() => {}))
      const { container } = render(<Calculator />)
      await typeNumber(user, 'First number', '6')
      await typeNumber(user, 'Second number', '3')
      await user.click(calculateButton())

      fireEvent.submit(container.querySelector('form')!)

      expect(calculateMock).toHaveBeenCalledTimes(1)
    })

    it('is disabled while waiting for the server', async () => {
      const user = userEvent.setup()
      let finish: (value: number) => void = () => {}
      calculateMock.mockReturnValue(
        new Promise<number>((resolve) => {
          finish = resolve
        }),
      )
      render(<Calculator />)

      await typeNumber(user, 'First number', '6')
      await typeNumber(user, 'Second number', '3')
      await user.click(calculateButton())

      const busyButton = screen.getByRole('button', { name: 'Calculating…' })
      expect(busyButton).toBeDisabled()

      finish(9)

      expect(await screen.findByRole('status')).toBeInTheDocument()
      expect(calculateButton()).toBeEnabled()
    })

    it('clears the result when the numbers change', async () => {
      const user = userEvent.setup()
      calculateMock.mockResolvedValue(9)
      render(<Calculator />)
      await typeNumber(user, 'First number', '6')
      await typeNumber(user, 'Second number', '3')
      await user.click(calculateButton())
      await screen.findByRole('status')

      await typeNumber(user, 'Second number', '1')

      expect(screen.queryByRole('status')).not.toBeInTheDocument()
    })

    it('clears the result when the operation changes', async () => {
      const user = userEvent.setup()
      calculateMock.mockResolvedValue(9)
      render(<Calculator />)
      await typeNumber(user, 'First number', '6')
      await typeNumber(user, 'Second number', '3')
      await user.click(calculateButton())
      await screen.findByRole('status')

      await selectOperation(user, 'Subtract')

      expect(screen.queryByRole('status')).not.toBeInTheDocument()
    })
  })

  describe('when the calculation fails', () => {
    it('shows the message sent by the backend', async () => {
      const user = userEvent.setup()
      calculateMock.mockRejectedValue(new ApiError('division by zero', 422))
      render(<Calculator />)

      await selectOperation(user, 'Divide')
      await typeNumber(user, 'First number', '5')
      await typeNumber(user, 'Second number', '0')
      await user.click(calculateButton())

      expect(await screen.findByRole('alert')).toHaveTextContent(
        'division by zero',
      )
      expect(screen.queryByRole('status')).not.toBeInTheDocument()
      expect(calculateButton()).toBeEnabled()
    })

    it('shows a generic message for an unexpected error', async () => {
      const user = userEvent.setup()
      calculateMock.mockRejectedValue(new Error('internal details'))
      render(<Calculator />)

      await typeNumber(user, 'First number', '6')
      await typeNumber(user, 'Second number', '3')
      await user.click(calculateButton())

      const alert = await screen.findByRole('alert')
      expect(alert).toHaveTextContent('Something went wrong')
      expect(alert).not.toHaveTextContent('internal details')
    })

    it('replaces an error with the next result', async () => {
      const user = userEvent.setup()
      calculateMock.mockRejectedValueOnce(
        new ApiError('Cannot reach the server'),
      )
      calculateMock.mockResolvedValueOnce(9)
      render(<Calculator />)
      await typeNumber(user, 'First number', '6')
      await typeNumber(user, 'Second number', '3')

      await user.click(calculateButton())
      expect(await screen.findByRole('alert')).toBeInTheDocument()

      await user.click(calculateButton())

      expect(await screen.findByRole('status')).toHaveTextContent('6 + 3 = 9')
      expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    })
  })
})
